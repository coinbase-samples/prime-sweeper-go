package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/coinbase-samples/prime-sdk-go"
	"github.com/coinbase-samples/prime-sweeper-go/utils"
	"os"
	"sort"
	"time"
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

	var allWallets []Wallet
	cursor := ""

	for {
		request := &prime.ListWalletsRequest{
			PortfolioId: client.Credentials.PortfolioId,
			Type:        "VAULT",
			Pagination: &prime.PaginationParams{
				Cursor:        cursor,
				Limit:         "1000",
				SortDirection: "ASC",
			},
		}

		ctx := context.Background()
		response, err := client.ListWallets(ctx, request)
		if err != nil {
			fmt.Printf("error listing wallets: %v\n", err)
			return
		}

		totalWalletCount := len(response.Wallets)

		for i, wallet := range response.Wallets {
			balanceResponse, err := client.GetWalletBalance(ctx, &prime.GetWalletBalanceRequest{
				PortfolioId: request.PortfolioId,
				Id:          wallet.Id,
			})
			if err != nil {
				fmt.Printf("error getting wallet balance for %s: %v\n", wallet.Name, err)
				continue
			}

			allWallets = append(allWallets, Wallet{
				Name:    wallet.Name,
				Id:      wallet.Id,
				Symbol:  wallet.Symbol,
				Balance: balanceResponse.Balance.WithdrawableAmount,
			})
			fmt.Printf("%d/%d: wallet %s (%s) written to csv\n", i+1, totalWalletCount, wallet.Name, wallet.Symbol)
		}

		if !response.HasNext() {
			break
		}

		cursor = response.Pagination.NextCursor
	}

	sort.Slice(allWallets, func(i, j int) bool {
		return allWallets[i].Symbol < allWallets[j].Symbol
	})

	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("cold_wallets_%s_%s.csv", client.Credentials.PortfolioId[:5], timestamp)

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
	for _, wallet := range allWallets {
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
