package strike

import (
	"encoding/json"
	"io/ioutil"
)

type balanceType struct {
	Currency string     `json:"currency"`
	Balance  amountType `json:"prepaidBalance"`
}

type bucketType struct {
	Used      amountType `json:"used,omitempty"`
	Remaining amountType `json:"remaining,omitempty"`
	Limit     amountType `json:"limit"`
}

type simpleType struct {
	Limit amountType `json:"limit"`
}

type limitType struct {
	Currency       string     `json:"currency"`
	PaymentsTotal  bucketType `json:"paymentsTotal,omitempty"`
	PaymentsSingle simpleType `json:"paymentsSingle,omitempty"`
	BTCDailyLimit  bucketType `json:"btcWithdrawalTotal1,omitempty"`
	BTCWeeklyLimit bucketType `json:"btcWithdrawalTotal2,omitempty"`
}

type balancesAndLimitsResponse struct {
	Balances []balanceType `json:"balances"`
	Limits   []limitType   `json:"limits"`
}

func (client StrikeClient) GetBalanceAndLimits() (balancesAndLimits balancesAndLimitsResponse, err error) {
	resp, err := client.sendGetRequest(balancesAndLimitsEndpoint)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return balancesAndLimits, err
	}
	json.Unmarshal(bodyBytes, &balancesAndLimits)

	return balancesAndLimits, err
}
