package strike

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Private Strike Endpoint and Methods
const (
	privateStrikeURL = "https://api.zaphq.io/api/v0.4/"

	balancesAndLimitsEndpoint = "user/info"

	withdrawEndpoint = "withdrawal/cryptoaddress" // POST

	exchangeEndpoint        = "exchange" // confirm with /:quoteId
	confirmExchangeEndpoint = "exchange/:quoteId"
	// {"exchangeType":"SELL","source":{"currency":"USD","amount":"10.00"},"currency":"BTC"}
	// {"quoteId":"101b81fe-c22c-44bd-bbc7-870f518fafa7","created":1665292489014,"validUntil":1665292498000,"instant":true,"source":{"amount":{"currency":"USD","amount":"10.00"},"fee":{"currency":"USD","amount":"0.00"},"total":{"currency":"USD","amount":"10.00"}},"target":{"amount":{"currency":"BTC","amount":"0.00051549"},"fee":{"currency":"BTC","amount":"0"},"total":{"currency":"BTC","amount":"0.00051549"}},"rate":{"amount":"19399.0184","sourceCurrency":"BTC","targetCurrency":"USD"}}

	historyEndpoint = "user/history"
	// {"items":[{"itemId":"9d0f7a8f-2f3b-4594-acea-7728baf280a7","total":{"currency":"USD","amount":"10.00"},"amount":{"currency
)

var privateEndpoints = []string{balancesAndLimitsEndpoint, withdrawEndpoint, exchangeEndpoint, confirmExchangeEndpoint, historyEndpoint}

const (
	publicStrikeURL = "https://api.strike.me/v1/"

	invoicesEndpoint     = "invoices"                // GET and POST
	quoteInvoiceEndpoint = "invoices/:quoteId/quote" // POST no payload

	ratesEndpoint = "rates/ticker" // GET
)

// Receives an amount defined in BTC, returns an invoice
// NOTE: The first time you run this, you need
func Withdraw(amount string) (invoice string) {
	return ""
}

func GetBalance() (string, error) {
	return "", nil
}

// Receives an amount defined in BTC, returns an invoice
func GetAddress(amount string) (invoice string) {
	// First we need to get price of BTC
	rates, err := getRates()
	if err != nil {
		log.Println(err)
		return ""
	}

	var price rateType
	for _, rate := range rates {
		if rate.TargetCurrency == "USD" && rate.SourceCurrency == "BTC" {
			price = rate
		}
	}

	// back calculate USD amount
	priceFloat, _ := strconv.ParseFloat(price.Amount, 64)
	amountFloat, _ := strconv.ParseFloat(amount, 64)
	USDfloat := amountFloat * priceFloat
	USDstring := fmt.Sprintf("%.2f", USDfloat)

	// Check if we will be able to withdraw our current balance + this rebalance
	// after taking our daily and weekly limits into consideration
	// Neither should be allowed to be exceeded, and leave $250 of buffer

	invoiceId, err := getInvoice("rebealanc for "+amount+" BTC", USDstring)
	if err != nil {
		log.Println(err)
		return ""
	}

	log.Println(invoiceId.InvoiceId)

	lnInvoice, err := getInvoiceQuote(invoiceId.InvoiceId)
	if err != nil {
		log.Println(err)
		return ""
	}

	return lnInvoice.Invoice
}

// Repurchase attempts to buy back all received BTC that are sitting as a USD balance without losses
func StrikeRepurchaser() (err error) {
	// First we grab our current Strike balance
	// Then we fetch our recents transactions (enough to sum up to our entire balance)
	// Then we see how much of our balance is from receives, and what BTC amount we sent to Strike
	// We check if we are able to repurchase the full BTC amount back. If not, keep looping and checking
	return nil
}

func sendGetRequest(endpoint string) (*http.Response, error) {
	client := &http.Client{}

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
		req.Header.Add("Authorization", "Bearer "+GoDotEnvVariable("STRIKE_JWT_TOKEN"))
	} else {
		req.Header.Add("Authorization", "Bearer "+GoDotEnvVariable("STRIKE_API_KEY"))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func sendPostRequest(endpoint string, payload interface{}) (*http.Response, error) {
	client := &http.Client{}

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

	log.Println(URL + endpoint)

	req, err := http.NewRequest("POST", URL+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	if contains(privateEndpoints, endpoint) {
		req.Header.Add("Authorization", "Bearer "+GoDotEnvVariable("STRIKE_JWT_TOKEN"))
	} else {
		req.Header.Add("Authorization", "Bearer "+GoDotEnvVariable("STRIKE_API_KEY"))
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return resp, nil
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

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
