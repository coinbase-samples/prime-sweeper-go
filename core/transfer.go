package core

import (
	"context"
	"fmt"
	"github.com/coinbase-samples/prime-sdk-go"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func findColdWalletIdForAsset(config *model.Config, asset string, walletType string) (string, error) {
	for _, wallet := range config.Wallets {
		if wallet.Asset == asset && wallet.Type == walletType {
			return wallet.ColdWalletId, nil
		}
	}
	return "", fmt.Errorf("cold wallet for asset '%s' of type '%s' not found", asset, walletType)
}

func findHotWalletIdForAsset(tradingWalletsMap map[string]WalletResponse, asset string) (string, error) {
	if walletResponse, exists := tradingWalletsMap[asset]; exists {
		return walletResponse.Id, nil
	}
	return "", fmt.Errorf("hot wallet for asset '%s' not found", asset)
}

func prepareTransferRequest(client *prime.Client,
	walletId string,
	balance *Balance,
	config *model.Config,
	direction model.TransferDirection,
) (prime.CreateWalletTransferRequest, error) {

	var (
		sourceWalletId      string
		destinationWalletId string
		err                 error
	)
	if direction == model.HotToCold {
		sourceWalletId = walletId
		destinationWalletId, err = findColdWalletIdForAsset(config, balance.Symbol, "cold_custody")
	} else if direction == model.ColdToHot {
		sourceWalletId = walletId
		destinationWalletId, err = findHotWalletIdForAsset(TradingWallets, balance.Symbol)
	}

	if err != nil {
		return prime.CreateWalletTransferRequest{}, err
	}

	amount := strconv.FormatFloat(balance.WithdrawableAmount, 'f', -1, 64)

	return prime.CreateWalletTransferRequest{
		PortfolioId:         client.Credentials.PortfolioId,
		SourceWalletId:      sourceWalletId,
		Symbol:              balance.Symbol,
		DestinationWalletId: destinationWalletId,
		IdempotencyKey:      uuid.New().String(),
		Amount:              amount,
	}, nil
}

func logAndTrackTransfer(response *prime.CreateWalletTransferResponse,
	config *model.Config,
	sourceWalletId string,
	destinationWalletId string,
	direction model.TransferDirection,
	log *zap.Logger,
	operationId string) {

	log.Info("initiated transfer",
		zap.String("amount", response.Amount),
		zap.String("symbol", response.Symbol),
		zap.String("source_wallet_id", sourceWalletId),
		zap.String("destination_wallet_id", destinationWalletId),
		zap.String("activity_id", response.ActivityId),
		zap.String("operation_id", operationId))

	if direction == model.ColdToHot {
		log.Info("cold transfer url", zap.String("transfer_url", response.ApprovalUrl), zap.String("operation_id", operationId))
	}

	go trackTransaction(response.ActivityId, log, config, response.ApprovalUrl, operationId)
}

func InitiateTransfers(walletsMap map[string]*Balance, config *model.Config, direction model.TransferDirection, log *zap.Logger, ruleName, operationId string) error {
	client, err := utils.GetClientFromEnv()
	if err != nil {
		log.Error("cannot get client from environment", zap.Error(err))
		return err
	}

	for walletId, balance := range walletsMap {
		ctx, cancel := utils.GetContextWithTimeout(config)

		request, err := prepareTransferRequest(client, walletId, balance, config, direction)
		if err != nil {
			log.Error("error preparing transfer request", zap.String("rule_name", ruleName), zap.String("operation_id", operationId), zap.Error(err))
			continue
		}

		response, err := client.CreateWalletTransfer(ctx, &request)
		cancel()
		if err != nil {
			log.Error("could not create transfer", zap.String("rule_name", ruleName), zap.String("operation_id", operationId), zap.Error(err))
			continue
		}

		logAndTrackTransfer(response, config, request.SourceWalletId, request.DestinationWalletId, direction, log, operationId)
	}

	return nil
}

func logTransactionStatus(client *prime.Client, ctx context.Context, transactionId, lastStatus, operationId string, log *zap.Logger) (string, error) {
	transactionResp, err := client.GetTransaction(ctx, &prime.GetTransactionRequest{
		PortfolioId:   client.Credentials.PortfolioId,
		TransactionId: transactionId,
	})
	if err != nil {
		log.Error("could not get transaction", zap.String("transaction_id", transactionId), zap.Error(err), zap.String("operation_id", operationId))
		return lastStatus, fmt.Errorf("could not get transaction for activity %s: %w", transactionId, err)
	}

	currentStatus := transactionResp.Transaction.Status
	if currentStatus != lastStatus {
		log.Info("transaction status updated", zap.String("transaction_id", transactionId), zap.String("status", currentStatus), zap.String("operation_id", operationId))
	}
	return currentStatus, nil
}

func trackTransaction(activityId string, log *zap.Logger, config *model.Config, approvalUrl, operationId string) error {
	client, err := utils.GetClientFromEnv()
	if err != nil {
		log.Error("cannot get client from environment", zap.Error(err))
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Daemon.TransferMonitorTimeoutDuration*time.Minute)
	defer cancel()

	activityResp, err := client.GetActivity(ctx, &prime.GetActivityRequest{
		PortfolioId: client.Credentials.PortfolioId,
		Id:          activityId,
	})
	if err != nil {
		log.Error("could not get activity", zap.String("activity_id", activityId), zap.String("operation_id", operationId), zap.Error(err))
		return fmt.Errorf("could not get activity: %w", err)
	}

	var lastStatus string
	transactionId := activityResp.Activity.ReferenceId

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded && approvalUrl != "" {
				log.Info("transaction tracking window exceeded, continue on Prime", zap.String("prime_url", approvalUrl), zap.String("operation_id", operationId))
			}
			return nil
		case <-time.After(config.Daemon.TransferMonitorFrequency * time.Second):
			var err error
			lastStatus, err = logTransactionStatus(client, ctx, transactionId, lastStatus, operationId, log)
			if err != nil {
				return err
			}

			if lastStatus == "TRANSACTION_DONE" || lastStatus == "TRANSACTION_REJECTED" || lastStatus == "TRANSACTION_FAILED" {
				return nil
			}
		}
	}
}
