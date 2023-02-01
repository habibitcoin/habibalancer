package lightning

import (
	"encoding/json"
	"io/ioutil"
)

type ListTransactionsResponse struct {
	Transactions []ListTransactionReponse `json:"transactions"`
}

type ListTransactionReponse struct {
	TxHash    string `json:"tx_hash"`
	TotalFees string `json:"total_fees"`
	Label     string `json:"label"`
}

func (client *LightningClient) ListTransactions() (txs ListTransactionsResponse, err error) {
	resp, err := client.sendGetRequest("v1/transactions")

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return txs, err
	}
	json.Unmarshal(bodyBytes, &txs)

	return txs, err
}
