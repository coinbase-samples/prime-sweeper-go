package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coinbase-samples/prime-sdk-go"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"net/http"
	"os"
	"time"
)

const defaultTimeoutDuration time.Duration = 7

func getTimeoutDuration(config *model.Config) time.Duration {
	if config.Daemon.ContextTimeoutDuration > 0 {
		return time.Duration(config.Daemon.ContextTimeoutDuration) * time.Second
	}

	return defaultTimeoutDuration * time.Second
}

func GetContextWithTimeout(config *model.Config) (context.Context, context.CancelFunc) {
	timeoutDuration := getTimeoutDuration(config)

	return context.WithTimeout(context.Background(), timeoutDuration)
}

func GetClientFromEnv() (*prime.Client, error) {
	credentials := &prime.Credentials{}
	if err := json.Unmarshal([]byte(os.Getenv("PRIME_CREDENTIALS")), credentials); err != nil {
		return nil, fmt.Errorf("cannot unmarshall credentials %w", err)
	}

	client := prime.NewClient(credentials, http.Client{})
	return client, nil
}

func LastStatusIsTerminal(status string) bool {
	return status == "TRANSACTION_DONE" || status == "TRANSACTION_REJECTED" || status == "TRANSACTION_FAILED"
}
