package safe_socket

import (
	"encoding/binary"
	"io"
)

const FrameHeaderSize = 4

func SendAll(socket io.Writer, bytes []byte) error {
	totalSent := 0
	for totalSent < len(bytes) {
		n, err := socket.Write(bytes[totalSent:])
		if err != nil {
			return err
		}
		totalSent += n
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	totalRecv := 0
	for totalRecv < size {
		n, err := socket.Read(buff[totalRecv:])
		if err != nil {
			return nil, err
		}
		totalRecv += n
	}
	return buff, nil
}

func SendFrame(socket io.Writer, payload []byte) error {
	header := make([]byte, FrameHeaderSize)
	binary.BigEndian.PutUint32(header, uint32(len(payload)))
	if err := SendAll(socket, header); err != nil {
		return err
	}
	return SendAll(socket, payload)
}

func RecvFrame(socket io.Reader) ([]byte, error) {
	header, err := RecvAll(socket, FrameHeaderSize)
	if err != nil {
		return nil, err
	}
	size := binary.BigEndian.Uint32(header)
	return RecvAll(socket, int(size))
}
