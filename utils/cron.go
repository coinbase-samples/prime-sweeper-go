package utils

import (
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func processTradingToColdCustody(log *zap.Logger, config *model.Config, rule model.Rule) {
	operationId := uuid.New().String()
	log.Info("checking for withdrawable trading balances to sweep to cold custody", zap.String("rule_name", rule.Name), zap.String("operation_id", operationId))

	assets := GetAssetsForRule(rule, config)
	filteredHotWallets := FilterHotWalletsByAssets(assets, TradingWallets)

	hotWalletIds := make([]string, 0, len(filteredHotWallets))
	for _, walletResponse := range filteredHotWallets {
		hotWalletIds = append(hotWalletIds, walletResponse.Id)
	}

	nonEmptyHotWallets, err := CollectWalletBalances(log, config, hotWalletIds)
	if err != nil {
		log.Error("failed to query hot wallet balances", zap.Error(err))
		return
	}

	if len(nonEmptyHotWallets) == 0 {
		log.Info("no hot wallets found with withdrawable balance", zap.String("rule_name", rule.Name), zap.String("operation_id", operationId))
		return
	}

	for walletId, balance := range nonEmptyHotWallets {
		log.Info("found hot wallet balance",
			zap.String("wallet_id", walletId),
			zap.String("symbol", balance.Symbol),
			zap.String("rule name", rule.Name),
			zap.Float64("withdrawable_amount", balance.WithdrawableAmount),
			zap.String("operation_id", operationId))
	}

	err = InitiateTransfers(nonEmptyHotWallets, config, HotToCold, log, operationId)
	if err != nil {
		log.Error("failed to initiate transfers", zap.String("operation_id", operationId), zap.Error(err))
	}
}

func processColdCustodyToTrading(log *zap.Logger, config *model.Config, rule model.Rule) {
	operationId := uuid.New().String()
	log.Info("checking for withdrawable cold custody balances to sweep to trading", zap.String("rule_name", rule.Name), zap.String("operation_id", operationId))

	coldWalletNamesInRule := rule.Wallets

	filteredColdWalletIds := FilterWalletsByName(coldWalletNamesInRule, config)

	nonEmptyColdWallets, err := CollectWalletBalances(log, config, filteredColdWalletIds)
	if err != nil {
		log.Error("Failed to query cold wallet balances", zap.Error(err))
		return
	}

	if len(nonEmptyColdWallets) == 0 {
		log.Info("no cold wallets found with withdrawable balance", zap.String("rule_name", rule.Name), zap.String("operation_id", operationId))
		return
	}

	for walletId, balance := range nonEmptyColdWallets {
		log.Info("found cold wallet balance",
			zap.String("wallet_id", walletId),
			zap.String("symbol", balance.Symbol),
			zap.String("rule name", rule.Name),
			zap.Float64("withdrawable_amount", balance.WithdrawableAmount),
			zap.String("operation_id", operationId))
	}

	err = InitiateTransfers(nonEmptyColdWallets, config, ColdToHot, log, operationId)
	if err != nil {
		log.Error("Failed to initiate transfers", zap.String("operation_id", operationId), zap.Error(err))
	}
}

func SetupAndRunCron(log *zap.Logger, config *model.Config) error {
	c := cron.New(cron.WithSeconds())

	for _, rule := range config.Rules {
		rule := rule
		_, cronErr := c.AddFunc(rule.Schedule, func() {
			switch rule.Direction {
			case "trading_to_cold_custody":
				processTradingToColdCustody(log, config, rule)
			case "cold_custody_to_trading":
				processColdCustodyToTrading(log, config, rule)
			}
		})
		if cronErr != nil {
			log.Error("Failed to schedule cron job for rule", zap.String("rule_name", rule.Name), zap.Error(cronErr))
			return cronErr
		}
	}

	c.Start()

	select {}
}
