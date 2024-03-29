package core

import (
	"fmt"
	"github.com/coinbase-samples/prime-sdk-go"
	"github.com/coinbase-samples/prime-sweeper-go/model"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type WalletResponse struct {
	Id     string `json:"id"`
	Symbol string `json:"symbol"`
}

var TradingWallets map[string]WalletResponse

var minTransactionAmount, _ = decimal.NewFromString("0.00000001")

type Balance struct {
	Id                 string          `json:"id"`
	Symbol             string          `json:"symbol"`
	WithdrawableAmount decimal.Decimal `json:"withdrawable_amount"`
}

func CollectTradingWallets(config *model.Config) (map[string]WalletResponse, error) {
	client, err := utils.GetClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("cannot get client from environment %w", err)
	}

	tradingWallets := make(map[string]WalletResponse)
	uniqueAssets := make(map[string]struct{})
	for _, wallet := range config.Wallets {
		uniqueAssets[wallet.Asset] = struct{}{}
	}

	ctx, cancel := utils.GetContextWithTimeout(config)
	defer cancel()

	for asset := range uniqueAssets {
		request := &prime.ListWalletsRequest{
			PortfolioId: client.Credentials.PortfolioId,
			Type:        "TRADING",
			Symbols:     []string{asset},
		}

		response, err := client.ListWallets(ctx, request)
		if err != nil {
			zap.L().Error("cannot list wallets for asset",
				zap.String("asset", asset),
				zap.Error(err),
			)
			return nil, err
		}

		found := false
		for _, wallet := range response.Wallets {
			if wallet.Symbol == asset {
				tradingWallets[asset] = WalletResponse{
					Id:     wallet.Id,
					Symbol: wallet.Symbol,
				}
				found = true
				break
			}
		}

		if !found {
			zap.L().Info("no trading wallet found for asset", zap.String("asset", asset))
		}
	}

	return tradingWallets, nil
}

func CollectWalletBalances(config *model.Config, walletIds []string) (map[string]*Balance, error) {
	nonEmptyWallets := make(map[string]*Balance)

	client, err := utils.GetClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("cannot get client from environment: %w", err)
	}

	for _, walletId := range walletIds {
		ctx, cancel := utils.GetContextWithTimeout(config)
		request := &prime.GetWalletBalanceRequest{
			PortfolioId: client.Credentials.PortfolioId,
			Id:          walletId,
		}

		response, err := client.GetWalletBalance(ctx, request)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("could not get balance for wallet ID %s: %v", walletId, err)
		}

		balance := response.Balance
		amount, err := decimal.NewFromString(balance.WithdrawableAmount)
		if err != nil {
			return nil, fmt.Errorf("could not parse amount for wallet ID %s: %v", walletId, err)
		}

		if amount.GreaterThan(minTransactionAmount) {
			nonEmptyWallets[walletId] = &Balance{
				Id:                 walletId,
				Symbol:             balance.Symbol,
				WithdrawableAmount: amount,
			}
		}
	}

	return nonEmptyWallets, nil
}

func GetAssetsForRule(rule model.Rule, config *model.Config) []string {
	var assets []string
	for _, walletName := range rule.Wallets {
		for _, wallet := range config.Wallets {
			if wallet.Name == walletName {
				assets = append(assets, wallet.Asset)
			}
		}
	}
	return assets
}

func FilterWalletsByAssets(assets []string, tradingWallets map[string]WalletResponse) map[string]WalletResponse {
	filteredWalletIds := make(map[string]WalletResponse)
	for asset, walletResponse := range tradingWallets {
		for _, ruleAsset := range assets {
			if asset == ruleAsset {
				filteredWalletIds[walletResponse.Id] = walletResponse
			}
		}
	}
	return filteredWalletIds
}

func FilterWalletsByName(walletNames []string, config *model.Config) []string {
	var filteredWalletIds []string
	for _, walletName := range walletNames {
		for _, wallet := range config.Wallets {
			if wallet.Name == walletName {
				filteredWalletIds = append(filteredWalletIds, wallet.WalletId)
			}
		}
	}
	return filteredWalletIds
}
