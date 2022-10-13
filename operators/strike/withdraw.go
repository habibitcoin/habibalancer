package strike

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type withdrawalPayload struct {
	Address string     `json:"address"`
	Amount  amountType `json:"amount"`
}

type withdrawalResponse struct {
	NewBalance amountType `json:"newPrepaidBalance"` // we want 0 amount after withdrawal
}

func (client StrikeClient) createWithdrawal(address string, BTCamount string) (success bool, err error) {
	var amount amountType
	amount.Currency = "BTC"
	amount.Amount = BTCamount
	resp, err := client.sendPostRequest(withdrawEndpoint, &withdrawalPayload{
		Address: address,
		Amount:  amount,
	})
	if err != nil {
		log.Println(err)
		return false, err
	}
	var withdrawalResp withdrawalResponse
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false, err
	}

	json.Unmarshal(bodyBytes, &withdrawalResp)
	if withdrawalResp.NewBalance.Amount != "0" {
		return false, err
	}

	return true, err
}
