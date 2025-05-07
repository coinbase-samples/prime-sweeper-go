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

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/coinbase-samples/prime-sdk-go/balances"
	"github.com/coinbase-samples/prime-sdk-go/client"
	"github.com/coinbase-samples/prime-sdk-go/model"
	"github.com/coinbase-samples/prime-sdk-go/wallets"
	primeWallets "github.com/coinbase-samples/prime-sdk-go/wallets"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
)

type Wallet struct {
	Name    string
	Id      string
	Symbol  string
	Balance string
}

func main() {
	client, err := utils.GetClientFromEnv()
	if err != nil {
		fmt.Printf("error getting client from environment: %v\n", err)
		return
	}

	var walletsWithBalances []Wallet

	wallets, err := loadAllWallets(client)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i, w := range wallets {

		balance, err := getBalance(client, client.Credentials().PortfolioId, w.Name, w.Id)
		if err != nil {
			fmt.Println(err)
			return
		}

		walletsWithBalances = append(walletsWithBalances, Wallet{
			Name:    w.Name,
			Id:      w.Id,
			Symbol:  w.Symbol,
			Balance: balance.WithdrawableAmount,
		})
		fmt.Printf("%d/%d: wallet %s (%s) written to csv\n", i+1, len(wallets), w.Name, w.Symbol)
	}

	sort.Slice(walletsWithBalances, func(i, j int) bool {
		return walletsWithBalances[i].Symbol < walletsWithBalances[j].Symbol
	})

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("cold_wallets_%s_%s.csv", client.Credentials().PortfolioId[:5], timestamp)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("error creating CSV file: %v\n", err)
		return
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"Name", "ID", "Symbol", "Balance"}); err != nil {
		fmt.Printf("error writing header to CSV file: %v\n", err)
		return
	}
	for _, wallet := range walletsWithBalances {
		if err := writer.Write([]string{
			wallet.Name,
			wallet.Id,
			wallet.Symbol,
			wallet.Balance,
		}); err != nil {
			fmt.Printf("error writing wallet %s to CSV file: %v\n", wallet.Name, err)
		}
	}

	fmt.Println("cold wallets have been successfully exported to csv in the local dir, sorted by symbol.")
}

func getBalance(client client.RestClient, portfolioId, walletName, walletId string) (*model.Balance, error) {

	svc := balances.NewBalancesService(client)

	ctx, cancel := context.WithTimeout(context.Background(), utils.DefaultTimeoutDurationInSeconds())
	defer cancel()

	response, err := svc.GetWalletBalance(
		ctx,
		&balances.GetWalletBalanceRequest{
			PortfolioId: portfolioId,
			Id:          walletId,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error getting wallet balance for %s: %w", walletName, err)
	}

	return response.Balance, nil
}

func loadAllWallets(client client.RestClient) ([]*model.Wallet, error) {

	svc := primeWallets.NewWalletsService(client)

	var wallets []*model.Wallet

	var cursor string
	for {
		req := &primeWallets.ListWalletsRequest{
			PortfolioId: client.Credentials().PortfolioId,
			Type:        "VAULT",
			Pagination: &model.PaginationParams{
				Cursor:        cursor,
				Limit:         1000,
				SortDirection: "ASC",
			},
		}

		w, c, err := loadWallets(svc, req)

		if err != nil {
			return nil, err
		}

		wallets = append(wallets, w...)

		if len(c) == 0 {
			break
		}

		cursor = c
	}

	return wallets, nil
}

func loadWallets(svc wallets.WalletsService, req *primeWallets.ListWalletsRequest) ([]*model.Wallet, string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), utils.DefaultTimeoutDurationInSeconds())
	defer cancel()

	response, err := svc.ListWallets(ctx, req)
	if err != nil {
		return nil, "", fmt.Errorf("error listing wallets %w", err)
	}

	return response.Wallets, response.Pagination.NextCursor, nil
}
