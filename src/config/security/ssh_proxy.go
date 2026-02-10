package security

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jedi4ever/addt/util"
)

// SSH Agent protocol message types
const (
	SSH_AGENTC_REQUEST_IDENTITIES = 11
	SSH_AGENT_IDENTITIES_ANSWER   = 12
	SSH_AGENTC_SIGN_REQUEST       = 13
	SSH_AGENT_SIGN_RESPONSE       = 14
	SSH_AGENT_FAILURE             = 5
)

// SSHProxyAgent creates a filtered SSH agent proxy
type SSHProxyAgent struct {
	upstreamSocket string
	proxySocket    string
	allowedKeys    []string // fingerprints or comments to allow
	listener       net.Listener
	mu             sync.Mutex
	running        bool
	allowedBlobs   map[string]bool   // cached key blobs that are allowed
	blobComments   map[string]string // maps blob to comment for audit logging
	useTCP         bool              // listen on TCP instead of Unix socket (macOS + podman)
	tcpPort        int               // TCP port when useTCP is true
}

// NewSSHProxyAgent creates a new SSH proxy agent
func NewSSHProxyAgent(upstreamSocket string, allowedKeys []string) (*SSHProxyAgent, error) {
	if upstreamSocket == "" {
		return nil, fmt.Errorf("upstream SSH_AUTH_SOCK not set")
	}

	// Create socket directory in addt home sockets/ so Podman machine can access it
	addtHome := util.GetAddtHome()
	if addtHome == "" {
		return nil, fmt.Errorf("failed to determine addt home directory")
	}

	socketsDir := filepath.Join(addtHome, "sockets")
	if err := os.MkdirAll(socketsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create sockets dir: %w", err)
	}

	// Create unique subdirectory for this proxy instance
	tmpDir, err := os.MkdirTemp(socketsDir, "ssh-proxy-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Set restrictive permissions on directory (owner only)
	if err := os.Chmod(tmpDir, 0700); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to set temp dir permissions: %w", err)
	}

	// Write PID file for cleanup to identify orphaned directories
	if err := WritePIDFile(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to write PID file: %w", err)
	}

	proxySocket := filepath.Join(tmpDir, "agent.sock")

	return &SSHProxyAgent{
		upstreamSocket: upstreamSocket,
		proxySocket:    proxySocket,
		allowedKeys:    allowedKeys,
		allowedBlobs:   make(map[string]bool),
		blobComments:   make(map[string]string),
	}, nil
}

// NewSSHProxyAgentTCP creates an SSH proxy agent that listens on TCP.
// Used on macOS where podman can't mount Unix sockets from the host.
func NewSSHProxyAgentTCP(upstreamSocket string, allowedKeys []string) (*SSHProxyAgent, error) {
	if upstreamSocket == "" {
		return nil, fmt.Errorf("upstream SSH_AUTH_SOCK not set")
	}

	return &SSHProxyAgent{
		upstreamSocket: upstreamSocket,
		allowedKeys:    allowedKeys,
		allowedBlobs:   make(map[string]bool),
		blobComments:   make(map[string]string),
		useTCP:         true,
	}, nil
}

// TCPPort returns the TCP port the proxy is listening on (only valid after Start with useTCP)
func (p *SSHProxyAgent) TCPPort() int {
	return p.tcpPort
}

// Start starts the proxy agent listener
func (p *SSHProxyAgent) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return nil
	}

	var listener net.Listener
	if p.useTCP {
		// TCP mode: must bind 0.0.0.0 because containers connect via
		// host.docker.internal which resolves to a non-loopback IP
		l, err := net.Listen("tcp", "0.0.0.0:0")
		if err != nil {
			return fmt.Errorf("failed to listen on TCP: %w", err)
		}
		p.tcpPort = l.Addr().(*net.TCPAddr).Port
		listener = l
	} else {
		l, err := net.Listen("unix", p.proxySocket)
		if err != nil {
			return fmt.Errorf("failed to listen on proxy socket: %w", err)
		}
		// Set restrictive permissions on socket (owner only)
		if err := os.Chmod(p.proxySocket, 0600); err != nil {
			l.Close()
			return fmt.Errorf("failed to set socket permissions: %w", err)
		}
		listener = l
	}

	p.listener = listener
	p.running = true

	// Pre-populate allowed blobs by connecting to upstream
	if err := p.populateAllowedBlobs(); err != nil {
		// Non-fatal, will filter on the fly
		fmt.Fprintf(os.Stderr, "Warning: could not pre-filter keys: %v\n", err)
	}

	go p.acceptLoop()

	return nil
}

// Stop stops the proxy agent
func (p *SSHProxyAgent) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.running = false
	if p.listener != nil {
		p.listener.Close()
	}

	// Clean up socket file and directory (Unix socket mode only)
	if !p.useTCP && p.proxySocket != "" {
		os.RemoveAll(filepath.Dir(p.proxySocket))
	}

	return nil
}

// SocketPath returns the path to the proxy socket
func (p *SSHProxyAgent) SocketPath() string {
	return p.proxySocket
}

func (p *SSHProxyAgent) acceptLoop() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			p.mu.Lock()
			running := p.running
			p.mu.Unlock()
			if !running {
				return
			}
			continue
		}
		go p.handleConnection(conn)
	}
}

func (p *SSHProxyAgent) handleConnection(client net.Conn) {
	defer client.Close()

	// Connect to upstream agent
	upstream, err := net.Dial("unix", p.upstreamSocket)
	if err != nil {
		return
	}
	defer upstream.Close()

	// Proxy messages bidirectionally with filtering
	var wg sync.WaitGroup
	wg.Add(2)

	// Client -> Upstream (filter sign requests)
	go func() {
		defer wg.Done()
		p.proxyClientToUpstream(client, upstream)
	}()

	// Upstream -> Client (filter identity responses)
	go func() {
		defer wg.Done()
		p.proxyUpstreamToClient(upstream, client)
	}()

	wg.Wait()
}

func (p *SSHProxyAgent) proxyClientToUpstream(client, upstream net.Conn) {
	for {
		msg, err := readAgentMessage(client)
		if err != nil {
			return
		}

		if len(msg) == 0 {
			continue
		}

		msgType := msg[0]

		// Filter sign requests - only allow for permitted keys
		if msgType == SSH_AGENTC_SIGN_REQUEST {
			allowed, keyComment := p.checkSignRequest(msg)
			if !allowed {
				// Send failure response
				LogSSHSign(keyComment, false, "key not in allowed list")
				writeAgentMessage(client, []byte{SSH_AGENT_FAILURE})
				continue
			}
			LogSSHSign(keyComment, true, "")
		}

		// Forward to upstream
		if err := writeAgentMessage(upstream, msg); err != nil {
			return
		}
	}
}

func (p *SSHProxyAgent) proxyUpstreamToClient(upstream, client net.Conn) {
	for {
		msg, err := readAgentMessage(upstream)
		if err != nil {
			return
		}

		if len(msg) == 0 {
			continue
		}

		msgType := msg[0]

		// Filter identity responses - only return allowed keys
		if msgType == SSH_AGENT_IDENTITIES_ANSWER {
			msg = p.filterIdentities(msg)
		}

		// Forward to client
		if err := writeAgentMessage(client, msg); err != nil {
			return
		}
	}
}

func (p *SSHProxyAgent) populateAllowedBlobs() error {
	conn, err := net.Dial("unix", p.upstreamSocket)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Request identities
	if err := writeAgentMessage(conn, []byte{SSH_AGENTC_REQUEST_IDENTITIES}); err != nil {
		return err
	}

	msg, err := readAgentMessage(conn)
	if err != nil {
		return err
	}

	if len(msg) == 0 || msg[0] != SSH_AGENT_IDENTITIES_ANSWER {
		return fmt.Errorf("unexpected response")
	}

	// Parse identities and cache allowed blobs with comments
	keys := parseIdentities(msg)
	for _, key := range keys {
		blobStr := string(key.blob)
		p.blobComments[blobStr] = key.comment
		if p.isKeyAllowed(key.comment, key.blob) {
			p.allowedBlobs[blobStr] = true
			LogSSHKeyAccess(key.comment, true)
		} else {
			LogSSHKeyAccess(key.comment, false)
		}
	}

	return nil
}

func (p *SSHProxyAgent) filterIdentities(msg []byte) []byte {
	keys := parseIdentities(msg)

	var allowed []sshKey
	for _, key := range keys {
		blobStr := string(key.blob)
		p.blobComments[blobStr] = key.comment
		if p.isKeyAllowed(key.comment, key.blob) {
			allowed = append(allowed, key)
			p.allowedBlobs[blobStr] = true
		}
	}

	return buildIdentitiesResponse(allowed)
}

func (p *SSHProxyAgent) isKeyAllowed(comment string, blob []byte) bool {
	// If no filter specified, allow all
	if len(p.allowedKeys) == 0 {
		return true
	}

	for _, filter := range p.allowedKeys {
		// Match by comment (filename, email, etc.)
		if strings.Contains(strings.ToLower(comment), strings.ToLower(filter)) {
			return true
		}
		// Match by fingerprint prefix
		if strings.HasPrefix(filter, "SHA256:") || strings.HasPrefix(filter, "MD5:") {
			// Would need to compute fingerprint - for now just match comment
			continue
		}
	}

	return false
}

// checkSignRequest checks if a sign request is allowed and returns the key comment for logging
func (p *SSHProxyAgent) checkSignRequest(msg []byte) (allowed bool, keyComment string) {
	if len(msg) < 5 {
		return false, "unknown"
	}

	// Parse key blob from sign request
	// Format: byte type, uint32 blob_len, blob, uint32 data_len, data, uint32 flags
	blobLen := binary.BigEndian.Uint32(msg[1:5])
	if len(msg) < int(5+blobLen) {
		return false, "unknown"
	}

	blob := msg[5 : 5+blobLen]
	blobStr := string(blob)

	// Find the comment for this blob
	p.mu.Lock()
	comment := p.blobComments[blobStr]
	isAllowed := p.allowedBlobs[blobStr]
	p.mu.Unlock()

	if comment == "" {
		comment = "unknown"
	}

	return isAllowed, comment
}

type sshKey struct {
	blob    []byte
	comment string
}

func parseIdentities(msg []byte) []sshKey {
	if len(msg) < 5 {
		return nil
	}

	// Skip message type byte
	data := msg[1:]

	// Read number of keys
	if len(data) < 4 {
		return nil
	}
	numKeys := binary.BigEndian.Uint32(data[:4])
	data = data[4:]

	var keys []sshKey
	for i := uint32(0); i < numKeys; i++ {
		// Read blob length
		if len(data) < 4 {
			break
		}
		blobLen := binary.BigEndian.Uint32(data[:4])
		data = data[4:]

		// Read blob
		if len(data) < int(blobLen) {
			break
		}
		blob := make([]byte, blobLen)
		copy(blob, data[:blobLen])
		data = data[blobLen:]

		// Read comment length
		if len(data) < 4 {
			break
		}
		commentLen := binary.BigEndian.Uint32(data[:4])
		data = data[4:]

		// Read comment
		if len(data) < int(commentLen) {
			break
		}
		comment := string(data[:commentLen])
		data = data[commentLen:]

		keys = append(keys, sshKey{blob: blob, comment: comment})
	}

	return keys
}

func buildIdentitiesResponse(keys []sshKey) []byte {
	var buf bytes.Buffer

	// Message type
	buf.WriteByte(SSH_AGENT_IDENTITIES_ANSWER)

	// Number of keys
	binary.Write(&buf, binary.BigEndian, uint32(len(keys)))

	for _, key := range keys {
		// Blob length and blob
		binary.Write(&buf, binary.BigEndian, uint32(len(key.blob)))
		buf.Write(key.blob)

		// Comment length and comment
		binary.Write(&buf, binary.BigEndian, uint32(len(key.comment)))
		buf.WriteString(key.comment)
	}

	return buf.Bytes()
}

func readAgentMessage(conn net.Conn) ([]byte, error) {
	// Read 4-byte length
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return nil, err
	}

	msgLen := binary.BigEndian.Uint32(lenBuf)
	if msgLen > 256*1024 { // Sanity check: 256KB max
		return nil, fmt.Errorf("message too large: %d", msgLen)
	}

	// Read message
	msg := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func writeAgentMessage(conn net.Conn, msg []byte) error {
	// Write 4-byte length
	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(msg)))

	if _, err := conn.Write(lenBuf); err != nil {
		return err
	}

	// Write message
	if _, err := conn.Write(msg); err != nil {
		return err
	}

	return nil
}
