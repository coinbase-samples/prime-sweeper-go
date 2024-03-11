# Prime Sweeper

[![GoDoc](https://godoc.org/github.com/coinbase-samples/prime-sweeper-go?status.svg)](https://godoc.org/github.com/coinbase-samples/prime-sweeper-go)
[![Go Report Card](https://goreportcard.com/badge/coinbase-samples/prime-sweeper-go)](https://goreportcard.com/report/coinbase-samples/prime-sweeper-go)

## Overview

The Prime Sweeper is a rules-based reference application that automates onchain transactions, specifically the transfer of assets between [Coinbase Prime's](https://prime.coinbase.com) trading balances and cold storage wallets. This application utilizes Prime's [REST APIs](https://docs.cloud.coinbase.com/prime/docs/introduction). Example use cases include automatic sweeping from trading balance to cold custody at a recurring interval, or performing daily movements from cold custody to trading at the beginning of a trading day.

## License

The Prime Sweeper is free and open source and released under the [Apache License, Version 2.0](LICENSE).

The application and code are only available for demonstration purposes.

## Usage

Setup of the Prime Sweeper involves two steps:
1. Creating rules within config.yaml
2. Passing API credentials as an environment variable 

### Config setup
The Prime Sweeper requires the configuration of a config file to set the rules that will be processed by the Prime Sweeper. Begin by copying `config.example.yaml` to a new file called `config.yaml`. Then, begin to populate this config file with the relevant information that is required for the Prime Sweeper to run. 

The `config.yaml` has three top level parameters: `rules`, `wallets`, and `daemon`.

**Rules** define the creation and management of cron jobs. 

- `name`: string identifier for a given rule 
- `direction`: currently, the only two supported directions are `trading_to_cold_custody` and `cold_custody_to_trading`
- `description`: optional string summary for a given rule
- `schedule`: uses default cron syntax to determine run frequency
- `wallets`: cold wallets names (as defined in the `wallets` section) that are in scope for a given rule. This also implicitly determines which assets are in scope. 

For example, the following rule will perform hot to cold transfers every 30 seconds from BTC and ETH trading balances to the listed cold wallets: 

```
  - name: "constant_hot_sweep"
    direction: "trading_to_cold_custody"
    description: "Transfer from trading to cold custody at regular cadence"
    schedule: "*/30 * * * * *"
    wallets:
      - "ETH_cold"
      - "BTC_cold"
```

**Wallets** is a dictionary of cold wallets that may be used in rules. Because Coinbase Prime uses a single, universal ID for an asset's trading balance, the Prime Sweeper automatically collects trading balance IDs. However, due to Prime supporting many cold wallets per asset, the Prime Sweeper requires that wallets are defined manually here. To be clear, **only cold vault wallets should be added to this section**.

- `name`: string identifier for a given wallet; does not need to match Prime wallet name 
- `asset`: symbol for given wallet; must match symbol given by Prime API, e.g. `BTC`
- `description`: optional string identifier for a given wallet
- `type`: currently, the only supported type is `cold_custody`
- `cold-wallet-id`: UUID reported by [List Portfolio Wallets](https://docs.cloud.coinbase.com/prime/reference/primerestapi_getwallets)

For example, the two wallets referred to in the above rule may be defined as follows: 

```
- name: "BTC_cold"
    asset: "BTC"
    description: "main cold wallet for BTC"
    type: "cold_custody"
    cold-wallet-id: "wallet_uuid"
  - name: "ETH_cold"
    asset: "ETH"
    description: "main cold wallet for ETH"
    type: "cold_custody"
    wallet_id: "wallet_uuid"
```

Please note that you may include additional wallets here without having them included in rules. Only wallets that are defined in rules will be in scope for a given cron job.

Wallet IDs must be requested via the Prime API. The REST endpoint [List Portfolio Wallets](https://docs.cloud.coinbase.com/prime/reference/primerestapi_getwallets) should be used to get these values, are defined as `id` in the REST response. Example scripts for listing wallets are written in [Go](https://github.com/coinbase-samples/prime-cli) and [Python](https://github.com/coinbase-samples/prime-scripts-py/blob/main/REST/prime_list_wallets.py).

**Daemon** denotes the timeout duration for API requests in seconds. 

## API credentials 

You will need to pass an environment variable via your terminal called `PRIME_CREDENTIALS` with your API and portfolio information.

Coinbase Prime API credentials can be created in the Prime web console under Settings -> APIs. 

`PRIME_CREDENTIALS` should match the following format:

```
export PRIME_CREDENTIALS='{
"accessKey":"ACCESSKEY_HERE",
"passphrase":"PASSPHRASE_HERE",
"signingKey":"SIGNINGKEY_HERE",
"portfolioId":"PORTFOLIOID_HERE",
}'
```

Once these are set, you may run the Prime Sweeper from the project's root directory with `go run main.go`.
