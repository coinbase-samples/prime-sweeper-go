package core

import (
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func processTransfers(log *zap.Logger,
	config *model.Config,
	rule model.Rule,
	transferDetails model.TransferDetails) {

	log.Info("checking for withdrawable balances",
		zap.String("rule_name", rule.Name),
		zap.String("operation_id", transferDetails.OperationId))

	var walletIds []string
	if transferDetails.Direction == model.HotToCold {
		assets := GetAssetsForRule(rule, config)
		filteredWallets := FilterHotWalletsByAssets(assets, TradingWallets)
		for _, wallet := range filteredWallets {
			walletIds = append(walletIds, wallet.Id)
		}
	} else if transferDetails.Direction == model.ColdToHot {
		filteredWalletIds := FilterWalletsByName(transferDetails.WalletNames, config)
		walletIds = filteredWalletIds
	}

	nonEmptyWallets, err := CollectWalletBalances(config, walletIds)
	if err != nil {
		log.Error("failed to query wallet balances", zap.Error(err),
			zap.String("rule_name", rule.Name),
			zap.String("operation_id", transferDetails.OperationId))
		return
	}

	if len(nonEmptyWallets) == 0 {
		log.Info("no wallets found with withdrawable balance",
			zap.String("rule_name", rule.Name),
			zap.String("operation_id", transferDetails.OperationId))
		return
	}

	for walletId, balance := range nonEmptyWallets {
		log.Info("found wallet balance",
			zap.String("wallet_id", walletId),
			zap.String("symbol", balance.Symbol),
			zap.String("rule name", rule.Name),
			zap.Float64("withdrawable_amount", balance.WithdrawableAmount),
			zap.String("operation_id", transferDetails.OperationId))
	}

	if err = InitiateTransfers(nonEmptyWallets, config, transferDetails.Direction, log, rule.Name, transferDetails.OperationId); err != nil {
		log.Error("failed to initiate transfers", zap.String("operation_id", transferDetails.OperationId), zap.Error(err))
	}
}

func SetupAndRunCron(log *zap.Logger, config *model.Config) error {
	c := cron.New(cron.WithSeconds())

	for _, rule := range config.Rules {
		localRule := rule
		_, cronErr := c.AddFunc(localRule.Schedule, func() {
			transferDetails := model.TransferDetails{
				Direction:   model.TransferDirection(localRule.Direction),
				WalletNames: localRule.Wallets,
				OperationId: uuid.New().String(),
				RuleName:    localRule.Name,
			}
			processTransfers(log, config, localRule, transferDetails)
		})
		if cronErr != nil {
			log.Error("failed to schedule cron job for rule", zap.String("rule_name", localRule.Name), zap.Error(cronErr))
			return cronErr
		}
	}

	c.Start()

	select {}
}