package core

import (
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"go.uber.org/zap"
)

func ProcessTransfers(
	config *model.Config,
	rule model.Rule,
	transferDetails model.TransferDetails) {

	zap.L().Info("checking for withdrawable balances",
		zap.Any("rule", rule),
		zap.String("operation_id", transferDetails.OperationId),
	)

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
		zap.L().Error("failed to query wallet balances", zap.Error(err),
			zap.Any("rule", rule),
			zap.String("operation_id", transferDetails.OperationId),
		)
		return
	}

	if err = InitiateTransfers(nonEmptyWallets, config, transferDetails.Direction, rule, transferDetails.OperationId); err != nil {
		zap.L().Error("failed to initiate transfers",
			zap.Any("rule", rule),
			zap.String("operation_id", transferDetails.OperationId),
			zap.Error(err),
		)
	}
}
