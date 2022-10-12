package strike

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

type exchangePayload struct {
	Type           string     `json:"exchangeType"`
	SourceAmount   amountType `json:"source"`
	TargetCurrency string     `json:"currency"`
}

type summaryType struct {
	Amount amountType `json:"amount"`
	Fee    amountType `json:"fee"`
}

type exchangeResponse struct {
	QuoteId    string      `json:"quoteId"`
	ValidUntil int         `json:"validUntil"`
	Created    int         `json:"created"`
	USD        summaryType `json:"source"`
	BTC        summaryType `json:"target"`
	Rate       rateType    `json:"rate"`
}

type exchangeQuoteResponse struct {
	QuoteId string `json:"quoteId"`
	Result  string `json:"result"`
}

// Receives an amount defined in USD, returns a quote.
func (client StrikeClient) exchange(sourceAmount string, sourceCurrency string, targetCurrency string) (quote exchangeResponse, err error) {
	var amount amountType
	amount.Currency = sourceCurrency
	amount.Amount = sourceAmount
	resp, err := client.sendPostRequest(exchangeEndpoint, &exchangePayload{
		Type:           "SELL",
		TargetCurrency: targetCurrency,
		SourceAmount:   amount,
	})
	if err != nil {
		log.Println(err)
		return quote, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return quote, err
	}
	json.Unmarshal(bodyBytes, &quote)

	return quote, nil
}

func (client StrikeClient) confirmExchange(quoteId string) (success bool, err error) {
	var quoteResponse exchangeQuoteResponse
	endpoint := strings.Replace(confirmExchangeEndpoint, ":quoteId", quoteId, 1)
	resp, err := client.sendPostRequest(endpoint, nil)
	if err != nil {
		log.Println(err)
		return false, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false, err
	}
	json.Unmarshal(bodyBytes, &quoteResponse)

	if quoteResponse.Result != "COMPLETED" {
		return false, err
	}

	return true, nil
}
