package deezy

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/habibitcoin/habibalancer/lightning"
	"github.com/joho/godotenv"
)

func IsChannelOpen() (status bool) {
	ChannelExists, err := lightning.ListChannels(GoDotEnvVariable("DEEZY_PEER"))
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

// use godot package to load/read the .env file and
// return the value of the key
func GoDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
