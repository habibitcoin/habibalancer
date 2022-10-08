
# Habibitcoin / Deezy.io Rebalancing Tool

This application is a tool to automate liquidity operations to earn fees from Deezy.io. Huge shoutout to @dannydiekroeger! Deezy.io operates a service where if you open a channel to Deezy, and move liquidity to Deezy's side, he'll pay you ~1000 ppm (variable) when you close your channel. Read more at [https://dannydeezy.notion.site/How-to-Earn-on-the-Lightning-Network-62c9ebee3403487bbbd936d75d59d6fc](https://dannydeezy.notion.site/How-to-Earn-on-the-Lightning-Network-62c9ebee3403487bbbd936d75d59d6fc)

At a high level, this application automates the following flow in an endless loop:
1. Checking if we have enough onchain funds to open a channel with Deezy
2. Continuously attempting liquidity operations
3. Closing our channel with Deezy once our local balance is exhausted

The application utilizes the LND REST API, and Chrome Driver to automate Kraken's Lightning Functionality (their API does not support lightning yet)

## Requirements
You will need:
- an LND node that you have REST access to
- a machine with Go 1.18 or higher installed
- an account with Kraken (in the future, will be adding Strike, NiceHash and others)

## Setup
You'll want to edit the values in .env.sample and rename the file to .env

If you've never used golang before, you'll want to install Go 1.18, then download this project, open the habibitcoin directory and run `go run server.go`

