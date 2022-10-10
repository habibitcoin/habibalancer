package strike

import (
	"encoding/json"
	"io/ioutil"
)

type historyResponse struct {
	Items []itemResponse `json:"items"`
}

type itemResponse struct {
	Id                  string      `json:"itemId"`
	CounterpartyAmounts summaryType `json:"counterpartyAmounts,omitempty"`
	Total               amountType  `json:"total"`
	Amount              amountType  `json:"amount"`
	Rate                rateType    `json:"rate"`
	Type                string      `json:"type"`             // TODO: Make this a typed field. Relevant types for this script are OrderReceive and ExchangeSell
	State               string      `json:"transactionState"` // TODO: Make this a typed field. Relevant type is COMPLETED for now.
	Description         string      `json:"description"`      // "rebealanc" is our "signature"
}

// Receives an amount defined in BTC, returns an invoice
// NOTE: The first time you run this, you need.
func getHistory() (history historyResponse, err error) {
	resp, err := sendGetRequest(historyEndpoint)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return history, err
	}
	json.Unmarshal(bodyBytes, &history)

	return history, err
}
