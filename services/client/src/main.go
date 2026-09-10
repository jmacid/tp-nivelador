package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"

	client "github.com/7574-sistemas-distribuidos/tp-nivelador/src/client"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
)

func loadConfig() (client.ClientConfig, error) {
	agencyId := os.Getenv("AGENCY_ID")
	if agencyId == "" {
		return client.ClientConfig{}, errors.New("AGENCY_ID environment variable is required")
	}

	serverHost := os.Getenv("SERVER_HOST")
	if serverHost == "" {
		return client.ClientConfig{}, errors.New("SERVER_HOST environment variable is required")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return client.ClientConfig{}, errors.New("SERVER_PORT environment variable is required")
	}

	batchSizeStr := os.Getenv("BATCH_SIZE")
	if batchSizeStr == "" {
		return client.ClientConfig{}, errors.New("BATCH_SIZE environment variable is required")
	}
	batchSize, err := strconv.Atoi(batchSizeStr)
	if err != nil || batchSize <= 0 {
		return client.ClientConfig{}, fmt.Errorf("BATCH_SIZE must be a positive integer, got %q", batchSizeStr)
	}

	inputFile := os.Getenv("INPUT_FILE")
	if inputFile == "" {
		return client.ClientConfig{}, errors.New("INPUT_FILE environment variable is required")
	}

	outputFile := os.Getenv("OUTPUT_FILE")
	if outputFile == "" {
		return client.ClientConfig{}, errors.New("OUTPUT_FILE environment variable is required")
	}

	return client.ClientConfig{
		ServerHost: serverHost,
		ServerPort: serverPort,
		AgencyId:   agencyId,
		BatchSize:  batchSize,
		InputFile:  inputFile,
		OutputFile: outputFile,
	}, nil
}

func waitForShutdownSignal() <-chan struct{} {
	shutdown := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("sigterm", logger.InProgress)
		close(shutdown)
	}()
	return shutdown
}

func run() int {
	config, err := loadConfig()
	if err != nil {
		logger.Error("load-config", logger.Fail, "err", err)
		return 1
	}

	shutdown := waitForShutdownSignal()

	c, err := client.NewClient(config, shutdown)
	if err != nil {
		if errors.Is(err, client.ErrShutdown) {
			logger.Info("client-new", logger.Success, "reason", "sigterm")
			return 0
		}
		logger.Error("client-new", logger.Fail, "err", err)
		return 1
	}

	if err := c.Run(); err != nil {
		if c.IsShuttingDown() {
			logger.Info("client-run", logger.Success, "reason", "sigterm")
			return 0
		}
		logger.Error("client-run", logger.Fail, "err", err)
		return 1
	}
	return 0
}

func main() {
	debug.SetGCPercent(20)
	os.Exit(run())
}
