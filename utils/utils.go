package utils

import (
	"context"
	"encoding/json"
	"github.com/coinbase-samples/prime-sdk-go"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

const defaultTimeoutDuration time.Duration = 7

func getTimeoutDuration(config *model.Config) time.Duration {
	if config.Daemon.TimeoutDuration > 0 {
		return time.Duration(config.Daemon.TimeoutDuration) * time.Second
	}

	return defaultTimeoutDuration * time.Second
}

func GetContextWithTimeout(config *model.Config) (context.Context, context.CancelFunc) {
	timeoutDuration := getTimeoutDuration(config)

	return context.WithTimeout(context.Background(), timeoutDuration)
}

func GetClientFromEnv(log *zap.Logger) (*prime.Client, error) {
	credentials := &prime.Credentials{}
	if err := json.Unmarshal([]byte(os.Getenv("PRIME_CREDENTIALS")), credentials); err != nil {
		log.Error("cannot unmarshal credentials: %w", zap.Error(err))
		return nil, err
	}

	client := prime.NewClient(credentials, http.Client{})
	return client, nil
}
