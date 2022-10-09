package lightning

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
)

type PayResponse struct {
	Destination     string
	PaymentHash     string
	NumSatoshis     string
	Timestamp       string
	Expiry          string
	Description     string
	DescriptionHash string
	FallbackAddr    string
	CltvExpiry      string
	PaymentAddr     byte
	NumMsat         string
}

func GetPayReq(payreq string) (payment PayResponse, err error) {
	resp, err := sendGetRequest("v1/payreq/" + payreq)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return payment, err
	}
	payment = PayResponse{}
	json.Unmarshal(bodyBytes, &payment)

	return payment, err
}

func GetPaymentRequestValid(paymentRequest string) bool {
	// First see if invoice exists
	resp, err := sendGetRequest("v1/payreq/" + paymentRequest)
	if err != nil {
		return false
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	bodyString := string(bodyBytes)

	if strings.Contains(bodyString, "err") {
		return false
	}

	return true
}

type PaymentRequestPayload struct {
	PaymentRequest  string   `json:"payment_request"`
	TimeoutSeconds  string   `json:"timeout_seconds"`
	OutgoingChanIds []string `json:"outgoing_chan_ids,omit_empty"`
	FeeLimitSat     string   `json:"fee_limit_sat"`
}

func SendPayReq(payreq string, feeLimitSat string) (string, error) {
	excludeDeezy := GoDotEnvVariable("EXCLUDE_DEEZY_FROM_LIQ_OPS")

	payload := &PaymentRequestPayload{
		PaymentRequest: payreq,
		TimeoutSeconds: GoDotEnvVariable("PAY_TIMEOUT_SECONDS"),
		FeeLimitSat:    feeLimitSat,
	}
	if excludeDeezy == "true" {
		ChannelExists, err := ListChannels(GoDotEnvVariable("DEEZY_PEER"))
		if err != nil {
			log.Println(err)
			return "", err
		} else if len(ChannelExists.Channels) == 0 {
			// no open deezy channels, so we don't need to exclude him explicitly
		} else {
			// Loop through each channel IDs and append to []string array
			allChannels, err := ListChannels("")
			if err != nil {
				log.Println(err)
				return "", err
			}
			for _, channel := range allChannels.Channels {
				if channel.Peer != GoDotEnvVariable("DEEZY_PEER") {
					payload.OutgoingChanIds = append(payload.OutgoingChanIds, channel.ChannelId)
				}
			}
		}
	}
	resp, err := sendPostRequestJSON("v2/router/send", payload)
	if err != nil {
		log.Println(err)
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil
}
