package utils

import (
	"context"
	"fmt"
	"github.com/coinbase-samples/prime-sdk-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"time"
)

const (
	HotToCold                      TransferDirection = "hot_to_cold"
	ColdToHot                      TransferDirection = "cold_to_hot"
	TransferMonitorCheckFrequency  time.Duration     = 10
	TransferMonitorTimeoutDuration time.Duration     = 15
)

type TransferDirection string

func InitiateTransfers(walletsMap map[string]*Balance, config *Config, direction TransferDirection, log *zap.Logger, operationId string) error {
	client, err := GetClientFromEnv(log)
	if err != nil {
		log.Error("cannot get client from environment", zap.Error(err))
		return err
	}

	ctx, cancel := GetContextWithTimeout(config)
	defer cancel()

	for walletId, balance := range walletsMap {
		var sourceWalletId, destinationWalletId string

		if direction == HotToCold {
			sourceWalletId = walletId
			destinationWalletId = findColdWalletIdForAsset(config, balance.Symbol, "cold_custody")
		} else if direction == ColdToHot {
			sourceWalletId = walletId
			destinationWalletId = findHotWalletIdForAsset(TradingWallets, balance.Symbol)
		}

		amount := strconv.FormatFloat(balance.WithdrawableAmount, 'f', -1, 64)

		request := &prime.CreateWalletTransferRequest{
			PortfolioId:         client.Credentials.PortfolioId,
			SourceWalletId:      sourceWalletId,
			Symbol:              balance.Symbol,
			DestinationWalletId: destinationWalletId,
			IdempotencyKey:      uuid.New().String(),
			Amount:              amount,
		}

		response, err := client.CreateWalletTransfer(ctx, request)
		if err != nil {
			log.Error("could not create transfer", zap.String("operation_id", operationId), zap.Error(err))
			return fmt.Errorf("could not create transfer %s: %v", walletId, err)
		}

		log.Info("initiated transfer",
			zap.String("amount", response.Amount),
			zap.String("symbol", balance.Symbol),
			zap.String("source_wallet_id", sourceWalletId),
			zap.String("destination_wallet_id", destinationWalletId),
			zap.String("activity_id", response.ActivityId),
			zap.String("operation_id", operationId))

		if direction == ColdToHot {
			log.Info("cold transfer url", zap.String("transfer_url", response.ApprovalUrl), zap.String("operation_id", operationId))

		}

		go func(activityId, approvalUrl string) {
			if err := trackTransaction(activityId, log, approvalUrl, operationId); err != nil {
				log.Error("could not track transaction", zap.String("activity_id", activityId), zap.String("operation_id", operationId), zap.Error(err))
			}
		}(response.ActivityId, response.ApprovalUrl)

	}
	return nil
}

func findColdWalletIdForAsset(config *Config, asset string, walletType string) string {
	for _, wallet := range config.Wallets {
		if wallet.Asset == asset && wallet.Type == walletType {
			return wallet.ColdWalletId
		}
	}
	return ""
}

func findHotWalletIdForAsset(tradingWalletsMap map[string]WalletResponse, asset string) string {
	if walletResponse, exists := tradingWalletsMap[asset]; exists {
		return walletResponse.Id
	}
	return ""
}

func trackTransaction(activityId string, log *zap.Logger, approvalUrl, operationId string) error {
	client, err := GetClientFromEnv(log)
	if err != nil {
		log.Error("cannot get client from environment", zap.Error(err))
		return err
	}

	var lastStatus string

	ctx, cancel := context.WithTimeout(context.Background(), TransferMonitorTimeoutDuration*time.Second)
	defer cancel()

	activityResp, err := client.GetActivity(ctx, &prime.GetActivityRequest{
		PortfolioId: client.Credentials.PortfolioId,
		Id:          activityId,
	})
	if err != nil {
		log.Error("could not get activity", zap.String("activity_id", activityId), zap.String("operation_id", operationId), zap.Error(err))
		return fmt.Errorf("could not get activity: %w", err)
	}

	transactionId := activityResp.Activity.ReferenceId

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				if approvalUrl != "" {
					log.Info("transaction tracking window exceeded, continue on Prime", zap.String("prime_url", approvalUrl), zap.String("operation_id", operationId))
				}
				return nil
			}
			log.Error("tracking transaction timed out", zap.String("activity_id", activityId), zap.String("operation_id", operationId))
			return fmt.Errorf("tracking transaction timed out")
		case <-time.After(TransferMonitorCheckFrequency * time.Second):
			transactionResp, err := client.GetTransaction(ctx, &prime.GetTransactionRequest{
				PortfolioId:   client.Credentials.PortfolioId,
				TransactionId: transactionId,
			})
			if err != nil {
				log.Error("could not get transaction", zap.String("transaction_id", transactionId), zap.String("activity_id", activityId), zap.Error(err), zap.String("operation_id", operationId))
				return fmt.Errorf("could not get transaction for activity %s: %w", activityId, err)
			}

			currentStatus := transactionResp.Transaction.Status
			if currentStatus != lastStatus {
				log.Info("transaction status updated", zap.String("transaction_id", transactionId), zap.String("status", currentStatus), zap.String("operation_id", operationId))
				lastStatus = currentStatus
			}

			if currentStatus == "TRANSACTION_DONE" || currentStatus == "TRANSACTION_REJECTED" || currentStatus == "TRANSACTION_FAILED" {
				return nil
			}
		}
	}
}
