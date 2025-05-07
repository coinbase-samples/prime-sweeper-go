/**
 * Copyright 2025-present Coinbase Global, Inc.
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

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/coinbase-samples/prime-sweeper-go/agent"
	"github.com/coinbase-samples/prime-sweeper-go/cmd/setup"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "prime-sweeper",
	Short: "Automate trading/vault asset movements on Coinbase Prime",
	RunE: func(cmd *cobra.Command, args []string) error {

		log, err := zap.NewProduction()
		if err != nil {
			return fmt.Errorf("cannot initialize logger: %w")
		}

		zap.ReplaceGlobals(log)
		defer log.Sync()

		sweeperAgent, err := agent.NewSweeperAgent("config.yaml")
		if err != nil {
			zap.L().Fatal("failed to initialize sweeper agent", zap.Error(err))
		}

		if err := sweeperAgent.Setup(); err != nil {
			zap.L().Fatal("failed to setup sweeper agent", zap.Error(err))
		}

		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

		if err := sweeperAgent.Run(stopChan); err != nil {
			zap.L().Fatal("error running sweeper agent", zap.Error(err))
		}

		zap.L().Info("Sweeper shut down gracefully.")

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(setup.Cmd)
}
