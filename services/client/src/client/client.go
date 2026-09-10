package client

import (
	"bufio"
	"encoding/binary"
	"net"
	"os"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

const ECHO_CLIENT_MESSAGE_AMOUNT = 3
const ECHO_CLIENT_MESSAGE_DELAY_MS = 1000

const MESSAGE_HEADER_SIZE = 4

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close()

	inFile, err := os.Open(client.config.InputFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.OpenFile(client.config.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		linea := scanner.Text()

		payload := []byte(linea)

		header := make([]byte, MESSAGE_HEADER_SIZE)
		binary.BigEndian.PutUint32(header, uint32(len(payload)))

		if err := safe_socket.SendAll(client.conn, header); err != nil {
			return err
		}
		if err := safe_socket.SendAll(client.conn, payload); err != nil {
			return err
		}

		respHeader, err := safe_socket.RecvAll(client.conn, MESSAGE_HEADER_SIZE)
		if err != nil {
			return err
		}
		respSize := binary.BigEndian.Uint32(respHeader)

		responseBuffer, err := safe_socket.RecvAll(client.conn, int(respSize))
		if err != nil {
			return err
		}

		outFile.WriteString(string(responseBuffer) + "\n")
	}

	return scanner.Err()
}
