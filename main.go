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

package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/coinbase-samples/prime-sweeper-go/agent"
	"go.uber.org/zap"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic("cannot initialize logger: " + err.Error())
	}
	zap.ReplaceGlobals(log)
	defer log.Sync()

	sweeperAgent, err := agent.NewSweeperAgent("config.yaml")
	if err != nil {
		zap.L().Error("failed to initialize sweeper agent", zap.Error(err))
		os.Exit(1)
	}

	if err := sweeperAgent.Setup(); err != nil {
		zap.L().Error("failed to setup sweeper agent", zap.Error(err))
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	if err := sweeperAgent.Run(stopChan); err != nil {
		zap.L().Error("error running sweeper agent", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info("Sweeper shut down gracefully.")
}
