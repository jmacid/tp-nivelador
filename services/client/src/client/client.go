package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const CONNECTION_ATTEMPTS_MAX = 3
const CONNECTION_ATTEMPS_DELAY_MS = 200

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

func parseBetLine(line string) (protocol.Bet, error) {
	fields := strings.Split(line, ",")
	if len(fields) != 5 {
		return protocol.Bet{}, fmt.Errorf("expected 5 fields in input line, got %d: %q", len(fields), line)
	}

	document, err := strconv.ParseUint(fields[2], 10, 32)
	if err != nil {
		return protocol.Bet{}, fmt.Errorf("invalid document in line %q: %w", line, err)
	}
	number, err := strconv.ParseUint(fields[4], 10, 32)
	if err != nil {
		return protocol.Bet{}, fmt.Errorf("invalid number in line %q: %w", line, err)
	}

	return protocol.Bet{
		FirstName: fields[0],
		LastName:  fields[1],
		Document:  uint32(document),
		Birthdate: fields[3],
		Number:    uint32(number),
	}, nil
}

func (client *Client) sendBets(inFile *os.File) error {
	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		bet, err := parseBetLine(scanner.Text())
		if err != nil {
			return err
		}

		payload, err := protocol.EncodeBet(bet)
		if err != nil {
			return err
		}
		if err := safe_socket.SendFrame(client.conn, payload); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return safe_socket.SendFrame(client.conn, protocol.EncodeDone())
}

func writeWinners(outFile *os.File, winners []protocol.Bet) error {
	for _, winner := range winners {
		line := fmt.Sprintf(
			"%s,%s,%d,%s,%d\n",
			winner.FirstName, winner.LastName, winner.Document, winner.Birthdate, winner.Number,
		)
		if _, err := outFile.WriteString(line); err != nil {
			return err
		}
	}
	return nil
}

func (client *Client) Run() error {
	const action = "run-bets"
	defer client.conn.Close()

	agencyId, err := strconv.ParseUint(client.config.AgencyId, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid AGENCY_ID %q: %w", client.config.AgencyId, err)
	}

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

	logger.Info(action, logger.InProgress, "agency-id", client.config.AgencyId)

	if err := safe_socket.SendFrame(client.conn, protocol.EncodeAgency(uint32(agencyId))); err != nil {
		return err
	}

	if err := client.sendBets(inFile); err != nil {
		return err
	}

	responsePayload, err := safe_socket.RecvFrame(client.conn)
	if err != nil {
		return err
	}
	winners, err := protocol.DecodeWinners(responsePayload)
	if err != nil {
		return err
	}

	if err := writeWinners(outFile, winners); err != nil {
		return err
	}

	logger.Info(action, logger.Success, "agency-id", client.config.AgencyId, "winners-amount", len(winners))
	return nil
}
