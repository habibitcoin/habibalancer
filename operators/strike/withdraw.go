package strike

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
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

type onchainSendPayload struct {
	Address  string `json:"address"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type onchainSendResponse struct {
	QuoteId        string     `json:"quoteId"`
	OnchainAddress string     `json:"address"`
	USDAmount      amountType `json:"amount"`
	USDTotal       amountType `json:"total"`
	Rate           rateType   `json:"rate"`
}

func (client StrikeClient) createOnchainSend(address string, USDamount string) (onchainSendQuote onchainSendResponse, err error) {
	var amount amountType
	amount.Currency = client.DefaultCurrency
	amount.Amount = USDamount
	resp, err := client.sendPostRequest(onchainSendEndpoint, &onchainSendPayload{
		Address:  address,
		Amount:   USDamount,
		Currency: client.DefaultCurrency,
	})
	if err != nil {
		log.Println(err)
		return onchainSendQuote, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return onchainSendQuote, err
	}

	json.Unmarshal(bodyBytes, &onchainSendQuote)

	return onchainSendQuote, err
}

type confirmOnchainSendResponse struct {
	OrderId string `json:"orderId"`
	Result  string `json:"result"`
}

func (client StrikeClient) confirmOnchainSend(quoteId string) (success bool, err error) {
	var quoteResponse confirmOnchainSendResponse
	endpoint := strings.Replace(confirmOnchainEndpoint, ":quoteId", quoteId, 1)
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
