package lightning

import (
	"encoding/json"
	"io/ioutil"
)

type BalanceResponse struct {
	TotalBalance       string `json:"total_balance"`
	ConfirmedBalance   string `json:"confirmed_balance"`
	UnconfirmedBalance string `json:"unconfirmed_balance"`
}

func GetBalance() (balances BalanceResponse, err error) {
	resp, err := sendGetRequest("v1/balance/blockchain")

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return balances, err
	}
	balances = BalanceResponse{}
	json.Unmarshal(bodyBytes, &balances)

	return balances, err
}
