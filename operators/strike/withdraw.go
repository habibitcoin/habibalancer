package strike

import (
	"io/ioutil"
)

type withdrawalPayload struct {
	Address string     `json:"address"`
	Amount  amountType `json:"amount"`
}

func createWithdrawal(address string, BTCamount string) (err error) {
	var amount amountType
	amount.Currency = "BTC"
	amount.Amount = BTCamount
	resp, err := sendPostRequest(withdrawEndpoint, &withdrawalPayload{
		Address: address,
		Amount:  amount,
	})

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return nil
}
