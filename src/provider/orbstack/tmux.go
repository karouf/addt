package orbstack

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/jedi4ever/addt/config/security"
	"github.com/jedi4ever/addt/util"
)

// tmuxProxy holds the proxy listener and connections
type tmuxProxy struct {
	listener       net.Listener
	upstreamSocket string
	proxyDir       string
	wg             sync.WaitGroup
	running        bool
	mu             sync.Mutex
	useTCP         bool
	tcpPort        int
}

// HandleTmuxForwarding handles tmux socket forwarding to the container.
// When enabled, it detects the active tmux socket and mounts it into the container,
// allowing the container to interact with the host's tmux session.
func (p *OrbStackProvider) HandleTmuxForwarding(enabled bool) []string {
	if !enabled {
		return nil
	}

	// Check if we're inside a tmux session
	tmuxEnv := os.Getenv("TMUX")
	if tmuxEnv == "" {
		// Not in a tmux session, nothing to forward
		return nil
	}

	// TMUX env format: /tmp/tmux-1000/default,12345,0
	// First part is the socket path
	parts := strings.Split(tmuxEnv, ",")
	if len(parts) < 1 {
		return nil
	}

	socketPath := parts[0]
	if socketPath == "" {
		return nil
	}

	// Verify socket exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil
	}

	// On macOS, Docker Desktop can't mount Unix sockets via virtiofs — use TCP bridge
	if runtime.GOOS == "darwin" {
		return p.handleTmuxForwardingTCP(socketPath, parts)
	}

	// Linux: create proxy socket and mount it directly
	proxyDir, proxySocket, err := p.createTmuxProxy(socketPath)
	if err != nil {
		fmt.Printf("Warning: failed to create tmux proxy: %v\n", err)
		return nil
	}

	newTmuxEnv := proxySocket
	if len(parts) > 1 {
		newTmuxEnv = proxySocket + "," + strings.Join(parts[1:], ",")
	}

	var args []string
	args = append(args, "-v", fmt.Sprintf("%s:%s", proxyDir, proxyDir))
	args = append(args, "-e", fmt.Sprintf("TMUX=%s", newTmuxEnv))

	if tmuxPane := os.Getenv("TMUX_PANE"); tmuxPane != "" {
		args = append(args, "-e", fmt.Sprintf("TMUX_PANE=%s", tmuxPane))
	}

	return args
}

// handleTmuxForwardingTCP uses TCP bridge for macOS + Docker Desktop.
func (p *OrbStackProvider) handleTmuxForwardingTCP(socketPath string, tmuxParts []string) []string {
	proxy, err := p.createTmuxProxyTCP(socketPath)
	if err != nil {
		fmt.Printf("Warning: failed to create tmux TCP proxy: %v\n", err)
		return nil
	}

	hostIP, err := getHostGatewayIP()
	if err != nil {
		fmt.Printf("Warning: could not detect host IP for tmux proxy: %v\n", err)
		proxy.Stop()
		return nil
	}

	var args []string
	args = append(args, "-e", fmt.Sprintf("ADDT_TMUX_PROXY_HOST=%s", hostIP))
	args = append(args, "-e", fmt.Sprintf("ADDT_TMUX_PROXY_PORT=%d", proxy.tcpPort))

	// Pass the remaining TMUX parts (pid, window) for reconstruction in the entrypoint
	if len(tmuxParts) > 1 {
		args = append(args, "-e", fmt.Sprintf("ADDT_TMUX_PARTS=%s", strings.Join(tmuxParts[1:], ",")))
	}

	if tmuxPane := os.Getenv("TMUX_PANE"); tmuxPane != "" {
		args = append(args, "-e", fmt.Sprintf("TMUX_PANE=%s", tmuxPane))
	}

	fmt.Println("Tmux forwarding active (TCP)")
	return args
}

// createTmuxProxyTCP creates a TCP-based tmux proxy for macOS.
func (p *OrbStackProvider) createTmuxProxyTCP(upstreamSocket string) (*tmuxProxy, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on TCP: %w", err)
	}

	proxy := &tmuxProxy{
		listener:       listener,
		upstreamSocket: upstreamSocket,
		running:        true,
		useTCP:         true,
		tcpPort:        listener.Addr().(*net.TCPAddr).Port,
	}

	p.tmuxProxy = proxy
	go proxy.acceptLoop()

	return proxy, nil
}

// createTmuxProxy creates a Unix socket proxy from addt sockets dir to the real tmux socket
func (p *OrbStackProvider) createTmuxProxy(upstreamSocket string) (string, string, error) {
	addtHome := util.GetAddtHome()
	if addtHome == "" {
		return "", "", fmt.Errorf("failed to determine addt home directory")
	}

	socketsDir := filepath.Join(addtHome, "sockets")
	if err := os.MkdirAll(socketsDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create sockets dir: %w", err)
	}

	proxyDir, err := os.MkdirTemp(socketsDir, "tmux-proxy-")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	if err := os.Chmod(proxyDir, 0700); err != nil {
		os.RemoveAll(proxyDir)
		return "", "", fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := security.WritePIDFile(proxyDir); err != nil {
		os.RemoveAll(proxyDir)
		return "", "", fmt.Errorf("failed to write PID file: %w", err)
	}

	socketName := filepath.Base(upstreamSocket)
	proxySocket := filepath.Join(proxyDir, socketName)

	listener, err := net.Listen("unix", proxySocket)
	if err != nil {
		os.RemoveAll(proxyDir)
		return "", "", fmt.Errorf("failed to create proxy socket: %w", err)
	}

	if err := os.Chmod(proxySocket, 0600); err != nil {
		listener.Close()
		os.RemoveAll(proxyDir)
		return "", "", fmt.Errorf("failed to set socket permissions: %w", err)
	}

	proxy := &tmuxProxy{
		listener:       listener,
		upstreamSocket: upstreamSocket,
		proxyDir:       proxyDir,
		running:        true,
	}

	p.tmuxProxy = proxy
	go proxy.acceptLoop()

	return proxyDir, proxySocket, nil
}

func (tp *tmuxProxy) acceptLoop() {
	for {
		conn, err := tp.listener.Accept()
		if err != nil {
			tp.mu.Lock()
			running := tp.running
			tp.mu.Unlock()
			if !running {
				return
			}
			continue
		}

		tp.wg.Add(1)
		go tp.handleConnection(conn)
	}
}

func (tp *tmuxProxy) handleConnection(client net.Conn) {
	defer tp.wg.Done()
	defer client.Close()

	upstream, err := net.Dial("unix", tp.upstreamSocket)
	if err != nil {
		return
	}
	defer upstream.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(upstream, client)
	}()

	go func() {
		defer wg.Done()
		io.Copy(client, upstream)
	}()

	wg.Wait()
}

func (tp *tmuxProxy) Stop() {
	tp.mu.Lock()
	tp.running = false
	tp.mu.Unlock()

	if tp.listener != nil {
		tp.listener.Close()
	}
	tp.wg.Wait()
	if !tp.useTCP && tp.proxyDir != "" {
		os.RemoveAll(tp.proxyDir)
	}
}
