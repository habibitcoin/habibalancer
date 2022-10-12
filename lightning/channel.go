package lightning

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
)

func (client *LightningClient) CreateChannel(peer string, amount int) (string, error) {
	peerHex, _ := hex.DecodeString(peer)
	peerUrl := base64.URLEncoding.EncodeToString(peerHex)

	resp, err := client.sendPostRequest("v1/channels", `{"node_pubkey":"`+peerUrl+`","sat_per_vbyte":"1","spend_unconfirmed":true,"private":false,"local_funding_amount":"`+strconv.Itoa(amount)+`"}`)
	if err != nil {
		log.Println(err)
		return "", err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, err
}

type ChannelsResponse struct {
	Channels []ChannelResponse `json:"channels"`
}

type ChannelResponse struct {
	Peer          string `json:"remote_pubkey"`
	ChannelId     string `json:"chan_id"`
	ChannelPoint  string `json:"channel_point"`
	LocalBalance  string `json:"local_balance"`
	RemoteBalance string `json:"remote_balance"`
}

func (client *LightningClient) ListChannels(peer string) (channels ChannelsResponse, err error) {
	peerHex, _ := hex.DecodeString(peer)
	peerUrl := base64.URLEncoding.EncodeToString(peerHex)
	prefix := ""
	if peer != "" {
		prefix = "?peer="
	}

	resp, err := client.sendGetRequest("v1/channels" + prefix + peerUrl)
	if err != nil {
		log.Println(err)
		return channels, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return channels, err
	}
	channels = ChannelsResponse{}
	if err := json.Unmarshal(bodyBytes, &channels); err != nil {
		log.Println(err)
		return channels, err
	}

	return channels, err
}

func (client *LightningClient) IsChannelOpen(peer string) (status bool) {
	ChannelExists, err := client.ListChannels(peer)
	if err != nil {
		log.Printf("Error listing channels in IsChannelOpen: %v", err)
		return false
	}

	return len(ChannelExists.Channels) != 0
}
