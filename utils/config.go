package utils

import (
	"fmt"
	"github.com/coinbase-samples/prime-sdk-go"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/go-yaml/yaml"
	"go.uber.org/zap"
	"os"
)

func ReadConfig(filename string) (*model.Config, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &model.Config{}

	if err = yaml.Unmarshal(bytes, config); err != nil {
		return nil, err
	}

	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfig(config *model.Config) error {
	if err := checkUniqueRuleNames(config); err != nil {
		return err
	}

	if err := checkRulesAndWallets(config); err != nil {
		return err
	}
	return validateColdWallets(config)
}

func checkUniqueRuleNames(config *model.Config) error {
	ruleNames := make(map[string]bool)
	for _, rule := range config.Rules {
		if _, exists := ruleNames[rule.Name]; exists {
			return fmt.Errorf("duplicate rule name: %s", rule.Name)
		}
		ruleNames[rule.Name] = true
	}
	return nil
}

func checkRulesAndWallets(config *model.Config) error {
	walletNames := make(map[string]bool)
	for _, rule := range config.Rules {
		if rule.Schedule == "" {
			return fmt.Errorf("schedule not specified for rule: %s", rule.Name)
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

func walletExists(walletName string, wallets []model.Wallet) bool {
	for _, w := range wallets {
		if w.Name == walletName {
			return true
		}
	}
	return false
}

func validateColdWallets(config *model.Config) error {
	client, err := GetClientFromEnv()
	if err != nil {
		return fmt.Errorf("cannot get client from environment %w", err)
	}

	for _, walletConfig := range config.Wallets {
		ctx, cancel := GetContextWithTimeout(config)

		request := &prime.GetWalletRequest{
			PortfolioId: client.Credentials.PortfolioId,
			Id:          walletConfig.WalletId,
		}

		response, err := client.GetWallet(ctx, request)
		cancel()
		if err != nil {
			zap.L().Error("cannot get wallet", zap.String("wallet", walletConfig.Name), zap.Error(err))
			return err
		}

		if response.Wallet.Symbol != walletConfig.Asset {
			return fmt.Errorf("asset mismatch for wallet '%s': expected '%s', got '%s'",
				walletConfig.Name,
				walletConfig.Asset,
				response.Wallet.Symbol,
			)
		}
	}

	return nil
}
