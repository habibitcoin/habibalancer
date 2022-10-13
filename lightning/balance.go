package lightning

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type BalanceResponse struct {
	TotalBalance       string `json:"total_balance"`
	ConfirmedBalance   string `json:"confirmed_balance"`
	UnconfirmedBalance string `json:"unconfirmed_balance"`
}

func (client *LightningClient) GetBalance() (balances BalanceResponse, err error) {
	resp, err := client.sendGetRequest("v1/balance/blockchain")
	if err != nil {
		log.Println(err)
		return balances, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return balances, err
	}
	balances = BalanceResponse{}
	if err := json.Unmarshal(bodyBytes, &balances); err != nil {
		log.Println(err)
		return balances, err
	}

	return balances, err
}
