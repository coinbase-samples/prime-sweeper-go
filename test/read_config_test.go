package test

import (
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReadConfig(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, filename, _, _ := runtime.Caller(0)
		dir := filepath.Dir(filename)

		expectedConfig := model.Config{
			Daemon: model.DaemonConfig{
				ContextTimeoutDuration:         15,
				TransferMonitorFrequency:       10,
				TransferMonitorTimeoutDuration: 300,
			},
			Rules: []model.Rule{
				{
					Name:        "daily_hot_sweep",
					Direction:   "trading_to_cold_custody",
					Description: "Transfer from trading to cold custody at specified time",
					Schedule:    "0 0 20 * * 1-5",
					Wallets:     []string{"ETH_cold"},
				},
			},
			Wallets: []model.Wallet{
				{
					Name:        "ETH_cold",
					Asset:       "ETH",
					Description: "main cold wallet for ETH",
					Type:        "cold_custody",
					WalletId:    "0ed06581-e121-4fe6-81df-1d5187432977",
				},
			},
		}

		configFilePath := filepath.Join(dir, "test_config.yaml")
		config, err := utils.ReadConfig(configFilePath)
		assert.NoError(t, err, "config should be loaded without errors")
		assert.Equal(t, expectedConfig, *config, "loaded config should match expected config")
	})

	t.Run("Failure", func(t *testing.T) {
		invalidConfigFilePath := "/path/to/nonexistent/config.yaml"
		_, err := utils.ReadConfig(invalidConfigFilePath)
		assert.Error(t, err, "an error was expected when attempting to read a non-existent or invalid config file")
	})
}
