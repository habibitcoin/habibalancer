package lightning

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/habibitcoin/habibalancer/configs"
)

type LightningClient struct {
	Client            *http.Client
	Host              string
	Macaroon          string
	Context           context.Context
	ExcludeDeezy      string
	TimeoutSeconds    string
	DeezyPeer         string
	DeezyClearnetHost string
	DeezyTorHost      string
	FeeRateSatsPerVb  string
}

// func NewLightningClient
func NewClient(ctx context.Context) (client LightningClient) {
	config := configs.GetConfig(ctx)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{
		Transport: tr,
	}
	client = LightningClient{
		Client:            httpClient,
		Host:              config.LNDHost,
		Macaroon:          loadMacaroon(ctx),
		ExcludeDeezy:      config.ExcludeDeezyFromLiqOps,
		DeezyClearnetHost: config.DeezyClearnetHost,
		DeezyTorHost:      config.DeezyTorHost,
		TimeoutSeconds:    config.PayTimeoutSeconds,
		FeeRateSatsPerVb:  config.FeeRateSatsPerVb,
		DeezyPeer:         config.DeezyPeer,
	}

	return client
}

func loadMacaroon(ctx context.Context) (macaroon string) {
	macaroonBytes, err := ioutil.ReadFile(configs.GetConfig(ctx).MacaroonLocation)
	if err != nil {
		log.Println("couldnt find or open macaroon")
		log.Println(err)
		return configs.GetConfig(ctx).Macaroon
	}

	macaroon = hex.EncodeToString(macaroonBytes)

	log.Println(macaroon)
	return macaroon
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
