/**
 * Copyright 2024-present Coinbase Global, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package core

import (
	"context"
	"fmt"
	"time"

	"github.com/coinbase-samples/prime-sdk-go/activities"
	"github.com/coinbase-samples/prime-sdk-go/client"
	"github.com/coinbase-samples/prime-sdk-go/transactions"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"go.uber.org/zap"
)

const maxWithdrawalGranularity int32 = 8

func findColdWalletIdForAsset(config *model.Config, asset string, walletType string) (string, error) {
	for _, wallet := range config.Wallets {
		if wallet.Asset == asset && wallet.Type == walletType {
			return wallet.WalletId, nil
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

func findWalletIdForAsset(config *model.Config, symbol string, direction model.TransferDirection) (string, error) {
	switch direction {
	case model.HotToCold:
		return findColdWalletIdForAsset(config, symbol, "cold_custody")
	case model.ColdToHot:
		return findHotWalletIdForAsset(TradingWallets, symbol)
	default:
		return "", fmt.Errorf("invalid transfer direction")
	}
}

func prepareTransferRequest(client client.RestClient,
	sourceWalletId string,
	balance *Balance,
	config *model.Config,
	direction model.TransferDirection,
) (*transactions.CreateWalletTransferRequest, error) {

	destinationWalletId, err := findWalletIdForAsset(config, balance.Symbol, direction)
	if err != nil {
		return nil, err
	}

	cappedAmount := balance.WithdrawableAmount.Truncate(maxWithdrawalGranularity)

	request := &transactions.CreateWalletTransferRequest{
		PortfolioId:         client.Credentials().PortfolioId,
		SourceWalletId:      sourceWalletId,
		Symbol:              balance.Symbol,
		DestinationWalletId: destinationWalletId,
		IdempotencyKey:      utils.NewUuid(),
		Amount:              cappedAmount.String(),
	}

	return request, nil
}

func logAndTrackTransfer(response *transactions.CreateWalletTransferResponse,
	config *model.Config,
	sourceWalletId string,
	destinationWalletId string,
	operationId string,
) {
	zap.L().Info("initiated transfer",
		zap.Any("response", response),
		zap.String("source_wallet_id", sourceWalletId),
		zap.String("destination_wallet_id", destinationWalletId),
		zap.String("operation_id", operationId),
	)

	go trackTransaction(response.ActivityId, config, response.ApprovalUrl, operationId)
}

func InitiateTransfers(
	walletsMap map[string]*Balance,
	config *model.Config,
	direction model.TransferDirection,
	rule model.Rule,
	operationId string,
) error {

	client, err := utils.GetClientFromEnv()
	if err != nil {
		zap.L().Error("cannot get client from environment", zap.Error(err))
		return err
	}

	svc := transactions.NewTransactionsService(client)

	for walletId, balance := range walletsMap {
		zap.L().Info("found wallet balance",
			zap.String("wallet_id", walletId),
			zap.Any("balance", balance),
			zap.Any("rule", rule),
			zap.String("operation_id", operationId),
		)

		ctx, cancel := utils.GetContextWithTimeout(config)
		request, err := prepareTransferRequest(client, walletId, balance, config, direction)
		if err != nil {
			zap.L().Error("error preparing transfer request",
				zap.Any("rule", rule),
				zap.String("wallet_id", walletId),
				zap.String("operation_id", operationId),
				zap.Error(err),
			)
			continue
		}

		response, err := svc.CreateWalletTransfer(ctx, request)
		cancel()
		if err != nil {
			zap.L().Error("could not create transfer",
				zap.Any("rule", rule),
				zap.String("wallet_id", walletId),
				zap.String("operation_id", operationId),
				zap.Error(err),
			)
			continue
		}

		logAndTrackTransfer(response, config, request.SourceWalletId, request.DestinationWalletId, operationId)
	}

	return nil
}

func logTransactionStatus(
	client client.RestClient,
	ctx context.Context,
	transactionId,
	lastStatus,
	operationId string,
) (string, error) {

	svc := transactions.NewTransactionsService(client)

	transactionResp, err := svc.GetTransaction(ctx, &transactions.GetTransactionRequest{
		PortfolioId:   client.Credentials().PortfolioId,
		TransactionId: transactionId,
	})
	if err != nil {
		zap.L().Error("could not get transaction",
			zap.String("transaction_id",
				transactionId),
			zap.Error(err),
			zap.String("operation_id", operationId),
		)
		return lastStatus, fmt.Errorf("could not get transaction for activity %s: %w", transactionId, err)
	}

	currentStatus := transactionResp.Transaction.Status
	if currentStatus != lastStatus {
		zap.L().Info("transaction status updated",
			zap.String("transaction_id", transactionId),
			zap.String("status", currentStatus),
			zap.String("operation_id", operationId),
		)
	}
	return currentStatus, nil
}

func trackTransaction(activityId string, config *model.Config, approvalUrl, operationId string) error {
	client, err := utils.GetClientFromEnv()
	if err != nil {
		zap.L().Error("cannot get client from environment", zap.Error(err))
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Daemon.TransferMonitorTimeoutDuration*time.Minute)
	defer cancel()

	activitiesSvc := activities.NewActivitiesService(client)

	activityResp, err := activitiesSvc.GetActivity(ctx, &activities.GetActivityRequest{
		PortfolioId: client.Credentials().PortfolioId,
		Id:          activityId,
	})
	if err != nil {
		zap.L().Error("could not get activity",
			zap.String("activity_id", activityId),
			zap.String("operation_id", operationId),
			zap.Error(err),
		)
		return fmt.Errorf("could not get activity: %w", err)
	}

	var lastStatus string
	transactionId := activityResp.Activity.ReferenceId

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded && approvalUrl != "" {
				zap.L().Info("transaction tracking window exceeded, continue on Prime",
					zap.String("prime_url", approvalUrl),
					zap.String("operation_id", operationId),
				)
			}
			return nil
		case <-time.After(config.Daemon.TransferMonitorFrequency * time.Second):
			var err error
			lastStatus, err = logTransactionStatus(client, ctx, transactionId, lastStatus, operationId)
			if err != nil {
				return err
			}

			if utils.LastStatusIsTerminal(lastStatus) {
				return nil
			}
		}
	}
}
