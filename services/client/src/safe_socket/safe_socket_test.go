package safe_socket

import (
	"bytes"
	"io"
	"testing"
)

type shortReader struct {
	data     []byte
	maxChunk int
}

func (r *shortReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := r.maxChunk
	if n > len(r.data) {
		n = len(r.data)
	}
	if n > len(p) {
		n = len(p)
	}
	copy(p, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

type shortWriter struct {
	written  []byte
	maxChunk int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	n := w.maxChunk
	if n > len(p) {
		n = len(p)
	}
	w.written = append(w.written, p[:n]...)
	return n, nil
}

func TestSendFrame_WritesLengthPrefixedPayloadDespiteShortWrites(t *testing.T) {
	writer := &shortWriter{maxChunk: 3}

	if err := SendFrame(writer, []byte("hello")); err != nil {
		t.Fatalf("SendFrame returned error: %v", err)
	}

	want := []byte{0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
	if !bytes.Equal(writer.written, want) {
		t.Fatalf("SendFrame wrote %v, want %v", writer.written, want)
	}
}

func TestRecvFrame_ReadsLengthPrefixedPayloadDespiteShortReads(t *testing.T) {
	reader := &shortReader{
		data:     []byte{0, 0, 0, 8, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'},
		maxChunk: 3,
	}

	payload, err := RecvFrame(reader)
	if err != nil {
		t.Fatalf("RecvFrame returned error: %v", err)
	}

	want := []byte("abcdefgh")
	if !bytes.Equal(payload, want) {
		t.Fatalf("RecvFrame() = %v, want %v", payload, want)
	}
}

func TestRecvFrame_ReturnsEmptyPayloadForZeroLengthFrame(t *testing.T) {
	reader := &shortReader{data: []byte{0, 0, 0, 0}, maxChunk: 3}

	payload, err := RecvFrame(reader)
	if err != nil {
		t.Fatalf("RecvFrame returned error: %v", err)
	}

	if len(payload) != 0 {
		t.Fatalf("RecvFrame() = %v, want empty", payload)
	}
}
