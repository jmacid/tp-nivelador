package safe_socket

import "io"

//TODO: Complete with a short-read/short-write tolerant implementation

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
