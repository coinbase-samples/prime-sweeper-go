package agent

import (
	"fmt"
	"github.com/coinbase-samples/prime-sweeper-go/core"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

type SweeperAgent struct {
	Config *model.Config
}

func NewSweeperAgent(configPath string) (*SweeperAgent, error) {
	config, err := utils.ReadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return &SweeperAgent{
		Config: config,
	}, nil
}

func (a *SweeperAgent) Setup() error {
	var err error
	core.TradingWallets, err = core.CollectTradingWallets(a.Config)
	if err != nil {
		return fmt.Errorf("cannot collect trading wallets: %w", err)
	}
	zap.L().Info("successfully collected trading wallet information.", zap.Any("TradingWallets", core.TradingWallets))

	return nil
}

func (a *SweeperAgent) Run() error {
	if err := core.SetupAndRunCron(a.Config); err != nil {
		return fmt.Errorf("failed to setup and run cron jobs: %w", err)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	<-stopChan
	zap.L().Info("shutting down Sweeper Agent...")
	return nil
}
