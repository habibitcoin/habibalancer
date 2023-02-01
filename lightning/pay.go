package lightning

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
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

type ListPaymentsResponse struct {
	Payments         []ListPaymentResponse `json:"payments"`
	FirstIndexOffset string                `json:"first_index_offset"`
	LastIndexOffset  string                `json:"last_index_offset"`
	TotalNumPayments string                `json:"total_num_payments"`
}

type ListPaymentResponse struct {
	PaymentHash     string  `json:"payment_hash"`
	Value           string  `json:"value"`
	CreationDate    string  `json:"creation_date"`
	Fee             string  `json:"fee"`
	PaymentPreimage string  `json:"payment_preimage"`
	ValueSat        string  `json:"value_sat"`
	ValueMsat       string  `json:"value_msat"`
	PaymentRequest  string  `json:"payment_request"`
	Status          string  `json:"status"`
	FeeSat          string  `json:"fee_sat"`
	FeeMsat         string  `json:"fee_msat"`
	CreationTimeNs  string  `json:"creation_time_ns"`
	Htlcs           []HtlcP `json:"htlcs"`
	PaymentIndex    string  `json:"payment_index"`
	FailureReason   string  `json:"failure_reason"`
}

type HtlcP struct {
	AttemptID     string      `json:"attempt_id"`
	Status        string      `json:"status"`
	Route         Route       `json:"route"`
	AttemptTimeNs string      `json:"attempt_time_ns"`
	ResolveTimeNs string      `json:"resolve_time_ns"`
	Failure       interface{} `json:"failure"`
	Preimage      string      `json:"preimage"`
}

type Route struct {
	TotalTimeLock int    `json:"total_time_lock"`
	TotalFees     string `json:"total_fees"`
	TotalAmt      string `json:"total_amt"`
	Hops          []Hop  `json:"hops"`
	TotalFeesMsat string `json:"total_fees_msat"`
	TotalAmtMsat  string `json:"total_amt_msat"`
}

type Hop struct {
	ChanID           string       `json:"chan_id"`
	ChanCapacity     string       `json:"chan_capacity"`
	AmtToForward     string       `json:"amt_to_forward"`
	Fee              string       `json:"fee"`
	Expiry           int          `json:"expiry"`
	AmtToForwardMsat string       `json:"amt_to_forward_msat"`
	FeeMsat          string       `json:"fee_msat"`
	PubKey           string       `json:"pub_key"`
	TlvPayload       bool         `json:"tlv_payload"`
	MppRecord        interface{}  `json:"mpp_record"`
	AmpRecord        interface{}  `json:"amp_record"`
	Metadata         string       `json:"metadata"`
	CustomRecords    CustomRecord `json:"custom_records"`
}

type CustomRecord struct {
	DeezyRecord string `json:"5492373485"`
}

func (client *LightningClient) ListPayments() (payment ListPaymentsResponse, err error) {
	resp, err := client.sendGetRequest("v1/payments?include_incomplete=false&reversed=true")

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return payment, err
	}
	json.Unmarshal(bodyBytes, &payment)

	return payment, err
}

func (client *LightningClient) GetPayReq(ctx context.Context, payreq string) (payment PayResponse, err error) {
	resp, err := client.sendGetRequest("v1/payreq/" + payreq)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return payment, err
	}
	payment = PayResponse{}
	json.Unmarshal(bodyBytes, &payment)

	return payment, err
}

func (client *LightningClient) GetPaymentRequestValid(paymentRequest string) bool {
	// First see if invoice exists
	resp, err := client.sendGetRequest("v1/payreq/" + paymentRequest)
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
	PaymentRequest    string            `json:"payment_request"`
	TimeoutSeconds    string            `json:"timeout_seconds"`
	OutgoingChanIds   []string          `json:"outgoing_chan_ids,omit_empty"`
	FeeLimitSat       string            `json:"fee_limit_sat"`
	DestCustomRecords map[uint64][]byte `json:"dest_custom_records,omit_empty"`
}

func (client *LightningClient) SendPayReq(payreq string, feeLimitSat string) (string, error) {
	excludeDeezy := client.ExcludeDeezy
	timeoutSeconds := client.TimeoutSeconds
	deezyPeer := client.DeezyPeer

	payload := &PaymentRequestPayload{
		PaymentRequest: payreq,
		TimeoutSeconds: timeoutSeconds,
		FeeLimitSat:    feeLimitSat,
	}

	m := make(map[uint64][]byte)

	recordID, err := strconv.ParseUint("5492373485", 10, 64)
	if err != nil {
		log.Println(err)
		return "", err
	}

	hexValue, err := hex.DecodeString(deezyPeer)
	if err != nil {
		log.Println(err)
		return "", err
	}

	m[recordID] = hexValue

	payload.DestCustomRecords = m
	if excludeDeezy == "true" {
		ChannelExists, err := client.ListChannels(deezyPeer)
		if err != nil {
			log.Println(err)
			return "", err
		} else if len(ChannelExists.Channels) == 0 {
			// no open deezy channels, so we don't need to exclude him explicitly
		} else {
			// Loop through each channel IDs and append to []string array
			allChannels, err := client.ListChannels("")
			if err != nil {
				log.Println(err)
				return "", err
			}
			for _, channel := range allChannels.Channels {
				if channel.Peer != deezyPeer {
					payload.OutgoingChanIds = append(payload.OutgoingChanIds, channel.ChannelId)
				}
			}
		}
	}
	resp, err := client.sendPostRequestJSON("v2/router/send", payload)
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
