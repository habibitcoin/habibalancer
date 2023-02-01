package lightning

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type FeeEstimateResponse struct {
	Regular int `json:"regular"`
}

func (client *LightningClient) CreateChannel(peer string, amount int) (string, error) {
	peerHex, _ := hex.DecodeString(peer)
	peerUrl := base64.URLEncoding.EncodeToString(peerHex)

	satsPerVb, err := strconv.Atoi(client.FeeRateSatsPerVb)
	if err != nil || satsPerVb == 0 {
		// falling back to fee estimation
		req, err := http.NewRequest("GET", "https://api.blockchain.info/mempool/fees", nil)
		if err != nil {
			log.Println(err)
			satsPerVb = 1
		}
		resp, err := client.Client.Do(req)
		if err != nil {
			log.Println(err)
			satsPerVb = 1
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			satsPerVb = 1
		}

		var feeEstimateResponse FeeEstimateResponse

		if err := json.Unmarshal(bodyBytes, &feeEstimateResponse); err != nil {
			log.Println(err)
			satsPerVb = 1
		}
		satsPerVb = feeEstimateResponse.Regular
	}

	resp, err := client.sendPostRequest("v1/channels", `{"node_pubkey":"`+peerUrl+`","sat_per_vbyte":"`+strconv.Itoa(satsPerVb)+`","spend_unconfirmed":true,"private":false,"local_funding_amount":"`+strconv.Itoa(amount)+`"}`)
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
	Capacity      string `json:"capacity"`
}

func (client *LightningClient) ListClosedChannels(peer string) (channelIds []string, openingTxs []string, largestChannelCapacitySats int, err error) {
	largestChannelCapacitySats = 0

	resp, err := client.sendGetRequest("v1/channels/closed?cooperative=true")
	if err != nil {
		log.Println(err)
		return channelIds, openingTxs, 0, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return channelIds, openingTxs, 0, err
	}

	channels := ChannelsResponse{}
	if err := json.Unmarshal(bodyBytes, &channels); err != nil {
		log.Println(err)
		return channelIds, openingTxs, 0, err
	}

	for _, closedChan := range channels.Channels {
		if closedChan.Peer == peer {
			channelIds = append(channelIds, closedChan.ChannelId)
			txChunk := strings.Split(closedChan.ChannelPoint, ":")
			openingTxs = append(openingTxs, txChunk[0])
			chanCap, _ := strconv.Atoi(closedChan.Capacity)
			if chanCap > largestChannelCapacitySats {
				largestChannelCapacitySats = chanCap
			}
		}
	}

	return channelIds, openingTxs, largestChannelCapacitySats, err
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
		log.Println(err)
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
	client.ListClosedChannels(peer)
	ChannelExists, err := client.ListChannels(peer)
	if err != nil {
		log.Printf("Error listing channels in IsChannelOpen: %v", err)
		return false
	}

	if len(ChannelExists.Channels) != 0 {
		return true
	}

	if peer == client.DeezyPeer {
		// ensure connection
		if connected := client.isDeezyConnected(); !connected {
			if ok := client.connectDeezyClearnet(); !ok {
				if ok = client.connectDeezyTor(); !ok {
					log.Println("Unable to verify connection to Deezy")
				}
			}
		}
	}

	return false
}

type PeersResponse struct {
	Peers []Peers `json:"peers"`
}

type Peers struct {
	Peer string `json:"pub_key"`
}

func (client *LightningClient) isDeezyConnected() (status bool) {
	resp, err := client.sendGetRequest("v1/peers")
	if err != nil {
		log.Println(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	peers := PeersResponse{}
	if err := json.Unmarshal(bodyBytes, &peers); err != nil {
		log.Println(err)
		return false
	}

	for _, peer := range peers.Peers {
		if peer.Peer == client.DeezyPeer {
			return true
		}
	}

	return false
}

type LightningAddress struct {
	Peer string `json:"pubkey"`
	Host string `json:"host"`
}

type ConnectPeerPayload struct {
	Address LightningAddress `json:"addr"`
	Perm    bool             `json:"perm"`
	Timeout string           `json:"timeout"`
}

func (client *LightningClient) connectDeezyClearnet() (status bool) {
	payload := &ConnectPeerPayload{
		Address: LightningAddress{
			Peer: client.DeezyPeer,
			Host: client.DeezyClearnetHost,
		},
		Perm:    true,
		Timeout: "60",
	}
	resp, err := client.sendPostRequestJSON("v1/peers", payload)
	if err != nil {
		log.Println(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	bodyString := string(bodyBytes)
	log.Println(bodyString)

	if len(bodyString) > 5 {
		return false
	}

	return true
}

func (client *LightningClient) connectDeezyTor() (status bool) {
	torPayload := &ConnectPeerPayload{
		Address: LightningAddress{
			Peer: client.DeezyPeer,
			Host: client.DeezyTorHost,
		},
		Perm:    true,
		Timeout: "60",
	}
	resp, err := client.sendPostRequestJSON("v1/peers", torPayload)
	if err != nil {
		log.Println(err)
		return false
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	bodyString := string(bodyBytes)
	log.Println(bodyString)

	return true
}
