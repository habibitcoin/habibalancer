package deezy

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"

	"github.com/habibitcoin/habibalancer/lightning"
)

func IsChannelOpen(peer string) (status bool) {
	ChannelExists, err := lightning.ListChannels(peer)
	if err != nil {
		return false
	}
	if len(ChannelExists.Channels) == 0 {
		return false
	}
	return true
}

// Closes a channel to Deezy.io when provided a channel point - returns response body as a string
func CloseChannel(chanPoint string) (string, error) {
	signature, err := lightning.SignMessage("close " + chanPoint)
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

func sendPostRequest(endpoint string, payload string) (*http.Response, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}
	var jsonStr = []byte(payload)

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
