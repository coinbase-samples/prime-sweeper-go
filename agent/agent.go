package agent

import (
	"fmt"
	"github.com/coinbase-samples/prime-sweeper-go/core"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type SweeperAgent struct {
	Config *model.Config
	Cron   *cron.Cron
}

func NewSweeperAgent(configPath string) (*SweeperAgent, error) {
	config, err := utils.ReadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &SweeperAgent{
		Config: config,
		Cron:   cron.New(cron.WithSeconds()),
	}, nil
}

func (a *SweeperAgent) Setup() error {
	var err error
	core.TradingWallets, err = core.CollectTradingWallets(a.Config)
	if err != nil {
		return fmt.Errorf("cannot collect trading wallets: %w", err)
	}
	zap.L().Info("successfully collected trading wallet information.",
		zap.Any("TradingWallets", core.TradingWallets),
	)

	return nil
}

func (a *SweeperAgent) Run() error {
	for _, rule := range a.Config.Rules {
		rule := rule
		_, err := a.Cron.AddFunc(rule.Schedule, func() {
			transferDetails := model.TransferDetails{
				Direction:   model.TransferDirection(rule.Direction),
				WalletNames: rule.Wallets,
				OperationId: uuid.New().String(),
				RuleName:    rule.Name,
			}
			core.ProcessTransfers(a.Config, rule, transferDetails)
		})
		if err != nil {
			zap.L().Error("failed to schedule cron job for rule", zap.Any("rule", rule), zap.Error(err))
			return err
		}
	}

	a.Cron.Start()

	return nil
}

func (a *SweeperAgent) Stop() {
	a.Cron.Stop()
	zap.L().Info("cron scheduler stopped, sweeper agent shutting down.")
}
