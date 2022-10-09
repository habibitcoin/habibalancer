package strike

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	"github.com/google/uuid"
)

type amountType struct {
	Currency string `json:"currency"`
	Amount   string `json:"amount"`
}

type InvoicePayload struct {
	CorrelationId string     `json:"correlationId"`
	Description   string     `json:"description"`
	Amount        amountType `json:"amount"`
}

type InvoiceResponse struct {
	InvoiceId string `json:"invoiceId"`
}

type InvoiceQuoteResponse struct {
	QuoteId        string     `json:"quoteId"`
	Description    string     `json:"description"`
	Invoice        string     `json:"lnInvoice"`
	OnchainAddress string     `json:"onchainAddress"`
	USDAmount      amountType `json:"targetAmount"`
	BTCAmount      amountType `json:"sourceAmount"`
	Rate           rateType   `json:"conversionRate"`
}

func getInvoice(description string, USDamount string) (invoice InvoiceResponse, err error) {
	var amount amountType
	amount.Currency = "USD"
	amount.Amount = USDamount
	resp, err := sendPostRequest(invoicesEndpoint, &InvoicePayload{
		CorrelationId: uuid.New().String(),
		Description:   description,
		Amount:        amount,
	})

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return invoice, err
	}
	json.Unmarshal(bodyBytes, &invoice)

	return invoice, err
}

func getInvoiceQuote(quoteId string) (invoice InvoiceQuoteResponse, err error) {
	endpoint := strings.Replace(quoteInvoiceEndpoint, ":quoteId", quoteId, 1)
	resp, err := sendPostRequest(endpoint, nil)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return invoice, err
	}
	json.Unmarshal(bodyBytes, &invoice)

	return invoice, err
}

type rateType struct {
	Amount         string `json:"amount"`
	SourceCurrency string `json:"sourceCurrency"`
	TargetCurrency string `json:"targetCurrency"`
}

type RatesResponse []rateType

func getRates() (rates RatesResponse, err error) {
	log.Println(ratesEndpoint)
	resp, err := sendGetRequest(ratesEndpoint)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return rates, err
	}
	json.Unmarshal(bodyBytes, &rates)

	return rates, err
}