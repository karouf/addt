package util

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestDownloadProgress(t *testing.T) {
	var buf bytes.Buffer
	dp := NewDownloadProgress(1000, "Testing")
	dp.writer = &buf

	dp.Update(500)
	output := buf.String()
	if !strings.Contains(output, "50%") {
		t.Errorf("Expected output to contain '50%%', got: %s", output)
	}
	if !strings.Contains(output, "Testing") {
		t.Errorf("Expected output to contain 'Testing', got: %s", output)
	}
}

func TestDownloadProgressUnknownSize(t *testing.T) {
	var buf bytes.Buffer
	dp := NewDownloadProgress(-1, "Testing")
	dp.writer = &buf

	dp.Update(1024)
	output := buf.String()
	if !strings.Contains(output, "1.0 KB") {
		t.Errorf("Expected output to contain '1.0 KB', got: %s", output)
	}
}

func TestProgressReader(t *testing.T) {
	data := make([]byte, 1000)
	reader := bytes.NewReader(data)

	pr := NewProgressReader(reader, 1000, "Testing")
	var buf bytes.Buffer
	pr.progress.writer = &buf

	// Read all data
	result, err := io.ReadAll(pr)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(result) != 1000 {
		t.Errorf("Expected to read 1000 bytes, got %d", len(result))
	}
	if pr.read != 1000 {
		t.Errorf("Expected pr.read to be 1000, got %d", pr.read)
	}
}

func TestSpinner(t *testing.T) {
	var buf bytes.Buffer
	s := NewSpinner("Testing spinner")
	s.writer = &buf

	s.Start()
	s.UpdateMessage("Updated message")
	s.Stop()

	// Spinner should have written something
	if buf.Len() == 0 {
		t.Error("Expected spinner to write output")
	}
}

func TestSpinnerWithSuccess(t *testing.T) {
	var buf bytes.Buffer
	s := NewSpinner("Testing")
	s.writer = &buf

	s.Start()
	s.StopWithSuccess("Completed successfully")

	output := buf.String()
	if !strings.Contains(output, SuccessIcon) {
		t.Errorf("Expected output to contain success icon, got: %s", output)
	}
	if !strings.Contains(output, "Completed successfully") {
		t.Errorf("Expected output to contain success message, got: %s", output)
	}
}

func TestProgressBar(t *testing.T) {
	var buf bytes.Buffer
	pb := NewProgressBar(100, "Testing")
	pb.writer = &buf

	pb.Update(50, "Halfway")
	output := buf.String()
	if !strings.Contains(output, "50%") {
		t.Errorf("Expected output to contain '50%%', got: %s", output)
	}

	buf.Reset()
	pb.Increment("")
	if pb.current != 51 {
		t.Errorf("Expected current to be 51, got %d", pb.current)
	}
}
