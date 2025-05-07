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

package agent

import (
	"fmt"
	"os"
	"sync"

	"github.com/coinbase-samples/prime-sweeper-go/core"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type SweeperAgent struct {
	config *model.Config
	cron   *cron.Cron
}

func NewSweeperAgent(configPath string) (*SweeperAgent, error) {
	config, err := utils.ReadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &SweeperAgent{
		config: config,
		cron:   cron.New(cron.WithSeconds()),
	}, nil
}

func (a *SweeperAgent) Setup() error {
	var err error
	core.TradingWallets, err = core.CollectTradingWallets(a.config)
	if err != nil {
		return fmt.Errorf("cannot collect trading wallets: %w", err)
	}
	zap.L().Info("successfully collected trading wallet information.",
		zap.Any("TradingWallets", core.TradingWallets),
	)

	return nil
}

func (a *SweeperAgent) Run(stopChan <-chan os.Signal) error {
	var wg sync.WaitGroup

	for _, rule := range a.config.Rules {
		rule := rule
		_, err := a.cron.AddFunc(rule.Schedule, func() {
			wg.Add(1)
			defer wg.Done()

			transferDetails := model.TransferDetails{
				Direction:   model.TransferDirection(rule.Direction),
				WalletNames: rule.Wallets,
				OperationId: utils.NewUuid(),
				RuleName:    rule.Name,
			}
			core.ProcessTransfers(a.config, rule, transferDetails)
		})
		if err != nil {
			zap.L().Error("failed to schedule cron job for rule", zap.Any("rule", rule), zap.Error(err))
			return err
		}
	}

	a.cron.Start()

	<-stopChan

	a.Stop()
	wg.Wait()

	return nil
}

func (a *SweeperAgent) Stop() {
	a.cron.Stop()
	zap.L().Info("cron scheduler stopped, waiting for all jobs to complete.")
}
