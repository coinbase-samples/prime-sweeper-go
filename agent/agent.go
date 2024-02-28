package agent

import (
	"fmt"
	"github.com/coinbase-samples/prime-sweeper-go/core"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"sync"
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

func (a *SweeperAgent) Run() error {
	var wg sync.WaitGroup

	for _, rule := range a.config.Rules {
		rule := rule
		_, err := a.cron.AddFunc(rule.Schedule, func() {
			wg.Add(1)
			defer wg.Done()

			transferDetails := model.TransferDetails{
				Direction:   model.TransferDirection(rule.Direction),
				WalletNames: rule.Wallets,
				OperationId: uuid.New().String(),
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
	go func() {
		defer wg.Wait()
	}()

	return nil
}

func (a *SweeperAgent) Stop() {
	a.cron.Stop()
	zap.L().Info("cron scheduler stopped, waiting for all jobs to complete.")
	zap.L().Info("all jobs completed, sweeper agent shutting down.")
}
