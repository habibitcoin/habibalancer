package deezy

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/habibitcoin/habibalancer/lightning"
)

// Closes a channel to Deezy.io when provided a channel point - returns response body as a string.
func CloseChannel(chanPoint string, lightningClient lightning.LightningClient) (string, error) {
	signature, err := lightningClient.SignMessage("close " + chanPoint)
	if err != nil {
		return "", err
	}

	resp, err := sendPostRequest("v1/earn/closechannel", `{"channel_point":"`+chanPoint+`","signature":"`+signature.Signature+`"}`)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, err
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func sendPostRequest(endpoint string, payload string) (*http.Response, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}
	jsonStr := []byte(payload)

	req, err := http.NewRequest("POST", "https://api.deezy.io/"+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
