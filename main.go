package main

import (
	"github.com/coinbase-samples/prime-sweeper-go/agent"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
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

	if err := sweeperAgent.Run(); err != nil {
		zap.L().Error("error running sweeper agent", zap.Error(err))
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	zap.L().Info("Shutting down Sweeper Agent...")
}
