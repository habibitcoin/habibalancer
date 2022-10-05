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

func SendPayReqE(payreq string, fee_limit_sat string) (err error) {
	resp, err := sendPostRequest("v2/router/send", `{"payment_request":"`+payreq+`","timeout_seconds":"120","fee_limit_sat":"`+fee_limit_sat+`"}`)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}

	bodyString := string(bodyBytes)
	log.Println(bodyString)

	return err
}

func SendPayReq(payreq string, fee_limit_sat string) (string, error) {
	resp, err := sendPostRequest("v2/router/send", `{"payment_request":"`+payreq+`","timeout_seconds":"120","fee_limit_sat":"`+fee_limit_sat+`"}`)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, nil
}
