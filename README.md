
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
- an account with [Kraken](https://kraken.app.link/f1qONfjA4tb) OR Strike (get $5 signup bonus to Strike if you use my referral code [https://invite.strike.me/NI73SY](https://invite.strike.me/NI73SY))
- (in the future, will be adding NiceHash and others)

## Setup
Easy Setup: download one of the binaries, run the file, and visit localhost:1323 to get started! Watch the video below.
Mirror at [https://vimeo.com/770879037](https://vimeo.com/770879037)
[![Watch the video](https://i.imgur.com/YX7uPMi.png)](https://www.youtube.com/watch?v=5jpLN6EskDw)



Advanced Setup: If you've never used golang before, you'll want to install Go 1.19, then download this project, open the habibitcoin directory and run `go run server.go`

## What It Looks Like
Once you get running, you can update/save your configurations, and hit "Start Earning" when you are ready to start.

![image](https://user-images.githubusercontent.com/114780316/195450982-0f3a4e8a-e7f8-4b31-b4ea-0ccbcb89b9c3.png)

