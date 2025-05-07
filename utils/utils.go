/**
 * Copyright 2024-present Coinbase Global, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/coinbase-samples/prime-sdk-go/client"
	"github.com/coinbase-samples/prime-sdk-go/credentials"
	primeUtils "github.com/coinbase-samples/prime-sdk-go/utils"
	"github.com/coinbase-samples/prime-sweeper-go/model"
)

const defaultTimeoutDuration time.Duration = 7

func getTimeoutDuration(config *model.Config) time.Duration {
	if config.Daemon.ContextTimeoutDuration > 0 {
		return time.Duration(config.Daemon.ContextTimeoutDuration) * time.Second
	}
	return DefaultTimeoutDurationInSeconds()
}

func GetContextWithTimeout(config *model.Config) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), getTimeoutDuration(config))
}

func DefaultTimeoutDurationInSeconds() time.Duration {
	return defaultTimeoutDuration * time.Second
}

func GetClientFromEnv() (client.RestClient, error) {
	credentials := &credentials.Credentials{}
	if err := json.Unmarshal([]byte(os.Getenv("PRIME_CREDENTIALS")), credentials); err != nil {
		return nil, fmt.Errorf("cannot unmarshall credentials %w", err)
	}

	httpClient, err := client.DefaultHttpClient()
	if err != nil {
		return nil, fmt.Errorf("cannot get default http client %w", err)
	}

	client := client.NewRestClient(credentials, httpClient)
	return client, nil
}

func LastStatusIsTerminal(status string) bool {
	return status == "TRANSACTION_DONE" || status == "TRANSACTION_REJECTED" || status == "TRANSACTION_FAILED"
}

func NewUuid() string {
	return primeUtils.NewUuid()
}
