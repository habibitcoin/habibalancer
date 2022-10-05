package lightning

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

func CreateChannel(peer string, amount int) (string, error) {
	peerHex, _ := hex.DecodeString(peer)
	peerUrl := base64.URLEncoding.EncodeToString(peerHex)

	resp, err := sendPostRequest("v1/channels", `{"node_pubkey":"`+peerUrl+`","sat_per_vbyte":"1","spend_unconfirmed":true,"private":false,"local_funding_amount":"`+strconv.Itoa(amount)+`"}`)

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
	ChannelPoint  string `json:"channel_point"`
	LocalBalance  string `json:"local_balance"`
	RemoteBalance string `json:"remote_balance"`
}

func ListChannels(peer string) (channels ChannelsResponse, err error) {
	peerHex, _ := hex.DecodeString(peer)
	peerUrl := base64.URLEncoding.EncodeToString(peerHex)

	resp, err := sendGetRequest("v1/channels?peer=" + peerUrl)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return channels, err
	}
	channels = ChannelsResponse{}
	json.Unmarshal(bodyBytes, &channels)

	return channels, err
}
