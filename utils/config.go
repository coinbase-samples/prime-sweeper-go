package utils

import (
	"fmt"
	"github.com/coinbase-samples/prime-sdk-go"
	"github.com/go-yaml/yaml"
	"go.uber.org/zap"
	"os"
)

type Config struct {
	Daemon  DaemonConfig `yaml:"daemon"`
	Rules   []Rule       `yaml:"rules"`
	Wallets []Wallet     `yaml:"wallets"`
}

type DaemonConfig struct {
	TimeoutDuration int `yaml:"timeoutDuration"`
}

type Rule struct {
	Direction   string   `yaml:"direction"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"` // Optional
	Schedule    string   `yaml:"schedule"`
	Wallets     []string `yaml:"wallets"`
}

type Wallet struct {
	Name         string `yaml:"name"`
	Asset        string `yaml:"asset"`
	Description  string `yaml:"description"` // Optional
	Type         string `yaml:"type"`
	ColdWalletId string `yaml:"cold-wallet-id"`
}

func ReadConfig(log *zap.Logger, filename string) (*Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	if err := validateConfig(log, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateConfig(log *zap.Logger, config *Config) error {
	if err := checkUniqueRuleNames(config); err != nil {
		return err
	}

	if err := checkRulesAndWallets(config); err != nil {
		return err
	}
	return validateColdWallets(log, config)
}

func checkUniqueRuleNames(config *Config) error {
	ruleNames := make(map[string]bool)
	for _, rule := range config.Rules {
		if _, exists := ruleNames[rule.Name]; exists {
			return fmt.Errorf("duplicate rule name: %s", rule.Name)
		}
		ruleNames[rule.Name] = true
	}
	return nil
}

func checkRulesAndWallets(config *Config) error {
	walletNames := make(map[string]bool)
	for _, rule := range config.Rules {
		if rule.Schedule == "" {
			return fmt.Errorf("schedule not specified for rule: %s", rule.Description)
		}
		for _, walletName := range rule.Wallets {
			if !walletExists(walletName, config.Wallets) {
				return fmt.Errorf("wallet '%s' in rule '%s' does not exist", walletName, rule.Description)
			}
		}
	}
	for _, wallet := range config.Wallets {
		if _, exists := walletNames[wallet.Name]; exists {
			return fmt.Errorf("duplicate wallet name: %s", wallet.Name)
		}
		walletNames[wallet.Name] = true
	}
	return nil
}

func walletExists(walletName string, wallets []Wallet) bool {
	for _, w := range wallets {
		if w.Name == walletName {
			return true
		}
	}
	return false
}

func validateColdWallets(log *zap.Logger, config *Config) error {
	client, err := GetClientFromEnv(log)
	if err != nil {
		log.Error("cannot get client from environment", zap.Error(err))
		return err
	}

	ctx, cancel := GetContextWithTimeout(config)
	defer cancel()

	for _, walletConfig := range config.Wallets {
		request := &prime.GetWalletRequest{
			PortfolioId: client.Credentials.PortfolioId,
			Id:          walletConfig.ColdWalletId,
		}

		response, err := client.GetWallet(ctx, request)
		if err != nil {
			log.Error("cannot get wallet", zap.String("wallet", walletConfig.Name), zap.Error(err))
			return err
		}

		if response.Wallet.Symbol != walletConfig.Asset {
			return fmt.Errorf("asset mismatch for wallet '%s': expected '%s', got '%s'", walletConfig.Name, walletConfig.Asset, response.Wallet.Symbol)
		}
	}

	return nil
}
