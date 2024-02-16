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
	defer log.Sync()

	sweeperAgent, err := agent.NewSweeperAgent(log, "config.yaml")
	if err != nil {
		log.Error("failed to initialize sweeper agent", zap.Error(err))
		os.Exit(1)
	}

	if err := sweeperAgent.Setup(); err != nil {
		log.Error("failed to setup sweeper agent", zap.Error(err))
		os.Exit(1)
	}

	if err := sweeperAgent.Run(); err != nil {
		log.Error("error running sweeper agent", zap.Error(err))
		os.Exit(1)
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-stopChan

	log.Info("Shutting down Sweeper Agent...")
}
