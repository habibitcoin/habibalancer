package strike

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/habibitcoin/habibalancer/configs"
	"github.com/habibitcoin/habibalancer/lightning"
	"github.com/sqweek/dialog"
)

// Private Strike Endpoint and Methods.
const (
	privateStrikeURL = "https://api.zaphq.io/api/v0.4/"

	balancesAndLimitsEndpoint = "user/info"

	withdrawEndpoint       = "withdrawal/cryptoaddress" // POST
	onchainSendEndpoint    = "send/onchain/quote"       // POST
	confirmOnchainEndpoint = "send/pay/:quoteId"        // POST

	exchangeEndpoint        = "exchange" // confirm with /:quoteId
	confirmExchangeEndpoint = "exchange/:quoteId"

	historyEndpoint = "user/history"
)

var (
	privateEndpoints = []string{balancesAndLimitsEndpoint, withdrawEndpoint, exchangeEndpoint, confirmExchangeEndpoint, historyEndpoint, onchainSendEndpoint, confirmOnchainEndpoint}
)

const (
	publicStrikeURL = "https://api.strike.me/v1/"

	invoicesEndpoint     = "invoices"                // GET and POST
	quoteInvoiceEndpoint = "invoices/:quoteId/quote" // POST no payload

	ratesEndpoint = "rates/ticker" // GET
)

type StrikeClient struct {
	Client          *http.Client
	ApiKey          string
	JwtToken        string
	DefaultCurrency string
	Context         context.Context
}

// func NewClient
func NewClient(ctx context.Context) (client StrikeClient) {
	httpClient := &http.Client{}
	config := configs.GetConfig(ctx)
	client = StrikeClient{
		Client:          httpClient,
		ApiKey:          config.StrikeApiKey,
		JwtToken:        config.StrikeJwtToken,
		DefaultCurrency: config.StrikeDefaultCurrency,
		Context:         ctx,
	}

	return client
}

// Receives an amount defined in BTC, returns success.
func (client StrikeClient) Withdraw(lightningClient lightning.LightningClient) (bool, error) {
	var (
		config = configs.GetConfig(client.Context)

		strikeWithdrawAmtXBTmin, _ = strconv.ParseFloat(config.StrikeWithdrawBtcMin, 64)
	)

	strikeBalanceStringXBT, err := client.GetBalance()
	if err != nil {
		return false, err
	}
	log.Println("Strike balance BTC")
	log.Println(strikeBalanceStringXBT)
	strikeBalanceFloatXBT, _ := strconv.ParseFloat(strikeBalanceStringXBT, 64)

	if strikeBalanceFloatXBT > strikeWithdrawAmtXBTmin {
		address, err := lightningClient.CreateAddress()
		if err != nil {
			log.Println("Error from LND creating new address for Strike withdrawal")
			return false, err
		}

		success, err := client.createWithdrawal(address, strikeBalanceStringXBT)
		if err != nil {
			return false, err
		}
		return success, nil
	}

	log.Println("Balance too low for withdrawal " + strikeBalanceStringXBT)
	return false, nil
}

func (client StrikeClient) GetBalance() (string, error) {
	success, err := client.GetBalanceAndLimits()
	if err != nil {
		return "", err
	}
	btcBalance := success.Balances[0]
	if btcBalance.Balance.Currency != "BTC" {
		btcBalance = success.Balances[1]
	}
	return btcBalance.Balance.Amount, nil
}

// Receives an amount defined in BTC, returns an invoice.
func (client StrikeClient) GetAddress(amount string) (invoice string) {
	var (
		config = configs.GetConfig(client.Context)

		strikeDailyLimitBufferUSD, _  = strconv.ParseFloat(config.StrikeDailyLimitBufferUsd, 64)
		strikeWeeklyLimitBufferUSD, _ = strconv.ParseFloat(config.StrikeWeeklyLimitBufferUsd, 64)
	)
	// First we need to get price of BTC
	rates, err := client.getRates()
	if err != nil {
		log.Println(err)
		return ""
	}

	var price rateType
	for _, rate := range rates {
		if rate.TargetCurrency == client.DefaultCurrency && rate.SourceCurrency == "BTC" {
			price = rate
		}
	}

	// back calculate USD amount
	priceFloat, _ := strconv.ParseFloat(price.Amount, 64)
	amountFloat, _ := strconv.ParseFloat(amount, 64)
	USDfloat := amountFloat * priceFloat
	USDstring := fmt.Sprintf("%.2f", USDfloat)

	// See if we have enough left in our limits to repurchase + withdraw
	balanceAndLimits, err := client.GetBalanceAndLimits()
	if err != nil {
		return ""
	}

	BTCinUSDlimits := balanceAndLimits.Limits[0]
	if BTCinUSDlimits.Currency != client.DefaultCurrency {
		BTCinUSDlimits = balanceAndLimits.Limits[1]
		if BTCinUSDlimits.Currency != client.DefaultCurrency {
			BTCinUSDlimits = balanceAndLimits.Limits[2]
		}
	}

	dailyBTCWithdrawalRemaining, _ := strconv.ParseFloat(BTCinUSDlimits.BTCDailyLimit.Remaining.Amount, 64)
	weeklyBTCWithdrawalRemaining, _ := strconv.ParseFloat(BTCinUSDlimits.BTCWeeklyLimit.Remaining.Amount, 64)

	if ((dailyBTCWithdrawalRemaining - USDfloat) < strikeDailyLimitBufferUSD) || ((weeklyBTCWithdrawalRemaining - USDfloat) < strikeWeeklyLimitBufferUSD) {
		log.Println("Not enough buffer remanining to send to Strike with BTC limits, trying USD limit")
		paymentsUSDWithdrawalRemaining, _ := strconv.ParseFloat(BTCinUSDlimits.PaymentsTotal.Remaining.Amount, 64)
		if (paymentsUSDWithdrawalRemaining - USDfloat) < strikeWeeklyLimitBufferUSD {
			log.Println("Not enough buffer remanining to send to Strike with USD limits")
			return ""
		}
	}

	// Check if we will be able to withdraw our current balance + this rebalance
	// after taking our daily and weekly limits into consideration
	// Neither should be allowed to be exceeded, and leave $250 of buffer

	invoiceId, err := client.getInvoice("rebealanc for "+amount+" BTC", USDstring)
	if err != nil {
		log.Println(err)
		return ""
	}

	log.Println(invoiceId.InvoiceId)

	lnInvoice, err := client.getInvoiceQuote(invoiceId.InvoiceId)
	if err != nil {
		log.Println(err)
		return ""
	}

	return lnInvoice.Invoice
}

// Repurchase attempts to buy back all received BTC that are sitting as a USD balance without losses.
func StrikeRepurchaser(ctx context.Context, onchainAddress string) (err error) {
	var (
		config                              = configs.GetConfig(ctx)
		client                              = NewClient(ctx)
		strikeRepurchaserCooldownSeconds, _ = strconv.Atoi(config.StrikeRepurchaseCooldownSeconds)
		strikeRepurchaserManualMode         = config.StrikeRepurchaserManualMode
	)
	firstRun := true
	for {
		if !firstRun {
			time.Sleep(time.Duration(strikeRepurchaserCooldownSeconds) * time.Second)
		}
		firstRun = false

		// First we fetch our recents transactions (looking back to our most recent BTC purchase, so we dont impact other activity)
		recentTransactions, err := client.getHistory()
		if err != nil {
			log.Println("Strike repurchase crashed getting recentTransactions")
			log.Println(err)
			return err
		}

		if client.DefaultCurrency == "USD" && 1 == 0 {
			// Then we need to buy back all the BTC from receives that have the description "rebealanc". Sum these amounts.
			var spendAmountUSD float64
			var buyBackAmountBTC float64
			for _, transaction := range recentTransactions.Items {
				if strings.Contains(transaction.Description, "rebealanc") && transaction.Type == "OrderReceive" && transaction.State == "COMPLETED" {
					amountFloat, _ := strconv.ParseFloat(transaction.Amount.Amount, 64)
					spendAmountUSD = spendAmountUSD + amountFloat
					rateFloat, _ := strconv.ParseFloat(transaction.Rate.Amount, 64)
					buyBackAmountBTC = buyBackAmountBTC + (amountFloat / rateFloat)
				} else if transaction.Type == "ExchangeSell" && transaction.State == "COMPLETED" {
					break
				} else {
					continue
				}
			}
			if spendAmountUSD > 0 {
				// We check if we are able to repurchase the full BTC amount back. If not, keep looping and checking
				spendAmountUSDString := fmt.Sprintf("%.2f", spendAmountUSD)
				createdQuote, err := client.exchange(spendAmountUSDString, client.DefaultCurrency, "BTC")
				if err != nil {
					log.Println("Error creating quote for repurchaser")
					log.Println(err)
					continue
				} else if createdQuote.USD.Fee.Amount != "0.00" {
					log.Println(createdQuote.USD.Fee.Amount)
					log.Println("Unexpected non-zero fee!")
					os.Exit(1)
				}

				quotedBTC, _ := strconv.ParseFloat(createdQuote.BTC.Amount.Amount, 64)
				if quotedBTC >= buyBackAmountBTC {
					if strikeRepurchaserManualMode == "true" {
						dialog.Message("Suitable Strike Price found! You should spend %v USD to buy %v BTC back", spendAmountUSDString, createdQuote.BTC.Amount.Amount).Title("Valid Strike Quote Found!").Info()
					} else {
						time.Sleep(1 * time.Second)
						success, err := client.confirmExchange(createdQuote.QuoteId)
						if err != nil || !success {
							log.Println("Error accepting quote on exchange for repurchaser")
							log.Println(err)
						}
					}
				} else {
					log.Println("We wanted " + fmt.Sprintf("%.8f", buyBackAmountBTC) + " BTC for " + fmt.Sprintf("%.2f", spendAmountUSD) + " USD but we would only get " + fmt.Sprintf("%.8f", quotedBTC) + " BTC")
				}
			} else {
				log.Println("Nothing to buy back!")
			}
		} else {
			// Then we need to send back all the BTC from receives that have the description "rebealanc". Sum these amounts.
			var spendAmountUSD float64
			var buyBackAmountBTC float64
			for _, transaction := range recentTransactions.Items {
				if strings.Contains(transaction.Description, "rebealanc") && transaction.Type == "OrderReceive" && transaction.State == "COMPLETED" {
					amountFloat, _ := strconv.ParseFloat(transaction.Amount.Amount, 64)
					spendAmountUSD = spendAmountUSD + amountFloat
					rateFloat, _ := strconv.ParseFloat(transaction.Rate.Amount, 64)
					buyBackAmountBTC = buyBackAmountBTC + (amountFloat / rateFloat)
				} else if transaction.Type == "Order" && transaction.State == "COMPLETED" {
					break
				} else {
					continue
				}
			}

			if spendAmountUSD > 0 {
				// We check if we are able to repurchase the full BTC amount back. If not, keep looping and checking
				var chunkAmount float64
				var amountSent float64
				chunkAmount = 950.00 // Comply with Strike max onchain send limits
				amountSent = 0.00
				for amountSent < spendAmountUSD {
					time.Sleep(time.Duration(strikeRepurchaserCooldownSeconds) * time.Second)
					toSpend := spendAmountUSD - amountSent
					if toSpend > chunkAmount {
						toSpend = chunkAmount
					}
					toSpendString := fmt.Sprintf("%.2f", toSpend)
					createdQuote, err := client.createOnchainSend(onchainAddress, toSpendString)
					if err != nil {
						log.Println("Error creating quote for repurchaser")
						log.Println(err)
						break
					}
					createdQuoteFloat, _ := strconv.ParseFloat(createdQuote.Rate.Amount, 64)
					createdAmountFloat, _ := strconv.ParseFloat(createdQuote.USDAmount.Amount, 64)
					createdTotalFloat, _ := strconv.ParseFloat(createdQuote.USDTotal.Amount, 64)
					btcReceived := createdAmountFloat / createdQuoteFloat
					trueRate := createdTotalFloat / btcReceived
					if trueRate < (spendAmountUSD / buyBackAmountBTC) {
						// rate is favorable, confirm
						confirmQuote, err := client.confirmOnchainSend(createdQuote.QuoteId)
						if err != nil || !confirmQuote {
							log.Println("Error confirming onChainSend for repurchaser")
							log.Println(err)
							break
						} else {
							amountSent = amountSent + toSpend
						}
					} else {
						log.Println("We wanted " + fmt.Sprintf("%.8f", spendAmountUSD/buyBackAmountBTC) + " BTC/USD but rate was higher at " + fmt.Sprintf("%.8f", (trueRate)) + " BTC/USD")
					}
				}
			} else {
				log.Println("Nothing to buy back!")
			}
		}
	}

	return nil
}

func (client StrikeClient) sendGetRequest(endpoint string) (*http.Response, error) {
	URL := publicStrikeURL
	if contains(privateEndpoints, endpoint) {
		URL = privateStrikeURL
	}

	log.Println(URL + endpoint)

	req, err := http.NewRequest("GET", URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if contains(privateEndpoints, endpoint) {
		req.Header.Add("Authorization", "Bearer "+client.JwtToken)
	} else {
		req.Header.Add("Authorization", "Bearer "+client.ApiKey)
	}

	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (client StrikeClient) sendPostRequest(endpoint string, payload interface{}) (*http.Response, error) {
	jsonStr, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling")
		log.Println(err)
	}
	log.Println(string(jsonStr))

	URL := publicStrikeURL
	if contains(privateEndpoints, endpoint) {
		URL = privateStrikeURL
	}
	if strings.Contains(endpoint, "exchange") {
		// addresses bug for detecting endpoints with string replacing
		URL = privateStrikeURL
	}
	if strings.Contains(endpoint, "send/pay") {
		// addresses bug for detecting endpoints with string replacing
		URL = privateStrikeURL
	}

	log.Println(URL + endpoint)

	req, err := http.NewRequest("POST", URL+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	if contains(privateEndpoints, endpoint) {
		req.Header.Add("Authorization", "Bearer "+client.JwtToken)
	} else {
		req.Header.Add("Authorization", "Bearer "+client.ApiKey)
	}
	resp, err := client.Client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return resp, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
