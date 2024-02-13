package main

import (
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"go.uber.org/zap"
	"os"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic("cannot initialize Zap logger: " + err.Error())
	}
	defer log.Sync()

	config, err := utils.ReadConfig(log, "config.yaml")
	if err != nil {
		log.Error("failed to read config", zap.Error(err))
		os.Exit(1)
	}

	utils.TradingWallets, err = utils.CollectTradingWallets(log, config)
	if err != nil {
		log.Error("cannot collect trading wallets", zap.Error(err))
		os.Exit(1)
	}
	log.Info("successfully collected trading wallet information.", zap.Any("TradingWallets", utils.TradingWallets))

	err = utils.SetupAndRunCron(log, config)
	if err != nil {
		log.Error("failed to setup and run cron jobs", zap.Error(err))
		os.Exit(1)
	}
}
