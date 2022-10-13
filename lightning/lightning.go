package lightning

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/habibitcoin/habibalancer/configs"
)

type LightningClient struct {
	Client   *http.Client
	Host     string
	Macaroon string
	Context  context.Context
}

// func NewLightningClient
func NewClient(ctx context.Context) (client LightningClient) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Transport: tr,
	}
	client = LightningClient{
		Client:   httpClient,
		Host:     configs.GetConfig(ctx).LNDHost,
		Macaroon: loadMacaroon(ctx),
	}

	return client
}

func loadMacaroon(ctx context.Context) (macaroon string) {
	file, err := os.Open(configs.GetConfig(ctx).MacaroonLocation)
	if err != nil {
		return configs.GetConfig(ctx).Macaroon
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanRunes)

	var finalResult []string
	var finalOriginal []string

	for scanner.Scan() {
		original := fmt.Sprintf("%s ", scanner.Text())

		finalOriginal = append(finalOriginal, original)

		hexstring := fmt.Sprintf("%x ", scanner.Text())

		finalResult = append(finalResult, hexstring)
	}

	return finalResult[0]
}

func (client *LightningClient) sendGetRequest(endpoint string) (*http.Response, error) {
	log.Println(client.Host + endpoint)
	req, err := http.NewRequest("GET", client.Host+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", client.Macaroon)
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (client *LightningClient) sendPostRequestJSON(endpoint string, payload interface{}) (*http.Response, error) {
	jsonStr, err := json.Marshal(payload)

	req, err := http.NewRequest("POST", client.Host+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", client.Macaroon)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return resp, nil
}

func (client *LightningClient) sendPostRequest(endpoint string, payload string) (*http.Response, error) {
	jsonStr := []byte(payload)

	req, err := http.NewRequest("POST", client.Host+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", client.Macaroon)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
