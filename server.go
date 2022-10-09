package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/habibitcoin/habibalancer/deezy"
	"github.com/habibitcoin/habibalancer/lightning"
	"github.com/habibitcoin/habibalancer/operators/kraken"
	"github.com/joho/godotenv"
)

var (
	deezyPeer         = GoDotEnvVariable("DEEZY_PEER")
	minLoopSize, _    = strconv.Atoi(GoDotEnvVariable("LOOP_SIZE_MIN_SAT"))
	localAmountMin, _ = strconv.Atoi(GoDotEnvVariable("LOCAL_AMOUNT_MIN_SAT"))

	krakenAmtXBTmin, _         = strconv.ParseFloat(GoDotEnvVariable("KRAKEN_OP_MIN_BTC"), 64)
	krakenAmtXBTmax, _         = strconv.ParseFloat(GoDotEnvVariable("KRAKEN_OP_MAX_BTC"), 64)
	krakenWithdrawAmtXBTmin, _ = strconv.ParseFloat(GoDotEnvVariable("KRAKEN_WITHDRAW_BTC_MIN"), 64)
	maxLiqFeePpm, _            = strconv.ParseFloat(GoDotEnvVariable("MAX_LIQ_FEE_PPM"), 64)
)

func main() {
	looper()
}

func looper() (err error) {
	for {
		// Step 1: Find if we have a channel opened with Deezy
		chanExists := deezy.IsChannelOpen()
		log.Println(chanExists)

		// Step 2:  If we do not have an open channel, see if we have enough money to open one
		if !chanExists {
			Balance, err := lightning.GetBalance()
			if err != nil {
				log.Println("Unexpected error fetching on-chain balance")
				log.Println(err)
			}
			log.Println("Onchain balance")
			log.Println(Balance)

			// Step 3: Open Channel to Danny
			totalBalance, _ := strconv.Atoi(Balance.TotalBalance)
			if totalBalance > 16500000 {
				totalBalance = 16500000 // dont use wumbo channels
			}
			if totalBalance > minLoopSize {
				log.Println("Opening channel to Deezy")
				resp, err := lightning.CreateChannel(deezyPeer, totalBalance-500000) // leave 500000 cushion
				if err != nil {
					log.Println("Error opening channel")
					log.Println(err)
					continue
				}
				log.Println("Channel Opened Successfully!")
				log.Println(resp)

			}
		} else {
			// Check if our open channel with Deezy's local balance is less than minimum close satoshis
			channels, err := lightning.ListChannels(deezyPeer)
			if err != nil {
				log.Println("Unexpected error fetching channels")
				log.Println(err)
				continue
			}
			if len(channels.Channels) > 0 {
				// If our local balance is less than the minimum, lets get paid!
				balanceInt, _ := strconv.Atoi(channels.Channels[0].LocalBalance)
				if balanceInt < localAmountMin {
					result, err := deezy.CloseChannel(channels.Channels[0].ChannelPoint)
					if err != nil {
						log.Println("Error getting paid from Deezy")
						log.Println(err)
						continue
					}
					log.Println("Deezy paid us!")
					log.Println(result)
				}
			}
		}

		// Step 4: Attempt sequence of liquidity operations, starting with sends
		/* General logic should follow:
		a. Check how much local liquidity I have on each channel
		b. Based on the amount, start attempting liq operations via:
			i. Kraken
			ii. NiceHash
			iii. etc
		c. Check again if local liquidity is acceptable and if liq op balances exceed deezyAmt, and exceed deezyAmt per channel rate (1000ppm)
		d. Send funds back to ourselves
		*/

		// STAY IN LOOP UNTIL BALANCE OF OPERATORS IS > LIQUIDITY OPERATION AMOUNT
		// Fetch Kraken LN Deposit Address
		krakenAmtXBTi := krakenAmtXBTmin + rand.Float64()*(krakenAmtXBTmax-krakenAmtXBTmin)
		krakenAmtXBT := fmt.Sprintf("%.5f", krakenAmtXBTi)
		krakenAmtXBTfee := fmt.Sprintf("%.0f", krakenAmtXBTi*maxLiqFeePpm*100) // fee is in satoshis, we want at least 50% profit
		lnInvoice := kraken.GetAddress(krakenAmtXBT)
		if lnInvoice == "" {
			continue
		}
		log.Println(lnInvoice)
		// Try to pay invoice
		for consecutiveErrors := 0; consecutiveErrors <= 10; consecutiveErrors++ {
			_, err = lightning.SendPayReq(lnInvoice, krakenAmtXBTfee)
			if err != nil {
				log.Println(err)
				if consecutiveErrors == 9 {
					time.Sleep(900 * time.Second)
					continue
				}
			}
			consecutiveErrors = 11
		}

		// Step 5: Withdraw funds from Kraken if we have enough money to begin a liq operation
		// Get our Kraken balance in XBT
		krakenBalanceStringXBT, err := kraken.GetBalance()
		if err != nil {
			continue
		}
		log.Println("Kraken balance XBT")
		log.Println(krakenBalanceStringXBT)
		krakenBalanceFloatXBT, _ := strconv.ParseFloat(krakenBalanceStringXBT, 64)

		// Get our onChain balance in SAT
		Balance, err := lightning.GetBalance()
		if err != nil {
			log.Println("Unexpected error fetching on-chain balance")
			log.Println(err)
		}
		log.Println("Onchain balance SAT")
		log.Println(Balance)

		totalOnChainBalance, _ := strconv.Atoi(Balance.TotalBalance)

		if (krakenBalanceFloatXBT*100000000+float64(totalOnChainBalance)) > float64(minLoopSize) && krakenBalanceFloatXBT > krakenWithdrawAmtXBTmin {
			// Try to withdraw all Kraken BTC because operator balance > liq amount
			result, err := kraken.Withdraw(krakenBalanceStringXBT)
			if err != nil {
				continue
			}
			fmt.Printf("Kraken withdrawal successful: %+v\n", result)
		}
	}

	return nil
}

// use godot package to load/read the .env file and
// return the value of the key
func GoDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
