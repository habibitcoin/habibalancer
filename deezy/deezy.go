package deezy

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

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

// Closes a channel to Deezy.io when provided a channel point - returns response body as a string.
func CalculateEarnings(lightningClient lightning.LightningClient) (int, error) {
	satsEarned := 0
	firstInvoice := "0"

	channelIDs, largestChannelCapacitySats, err := lightningClient.ListClosedChannels(lightningClient.DeezyPeer)
	if err != nil {
		return 0, err
	}

	invoicesReceived, err := lightningClient.GetInvoices()
	if err != nil {
		return 0, err
	}

	for _, invoice := range invoicesReceived.Invoices {
		if invoice.State == "SETTLED" && invoice.IsKeysend {
			for _, htlc := range invoice.Htlcs {
				if invoice.AmtMsats == htlc.AmtMsat && htlc.State == "SETTLED" && contains(channelIDs, htlc.ChanId) {
					amount, _ := strconv.Atoi(invoice.AmtSats)
					satsEarned = satsEarned + amount
					if firstInvoice == "0" {
						firstInvoice = invoice.SettleDate
					}
				}
			}
		}
	}

	firstInvoiceInt, _ := strconv.Atoi(firstInvoice)
	secondsEarning := float64(int(time.Now().Unix()) - firstInvoiceInt)
	daysEarning := secondsEarning / 86400
	normalizeYearFactor := 365 / daysEarning
	normalizedEarnings := float64(satsEarned) * normalizeYearFactor
	projectedAPY := (normalizedEarnings / float64(largestChannelCapacitySats)) * 100
	log.Println("Deezy has paid us sats to date!")
	log.Println(satsEarned)
	log.Println("Estimated working capital in sats:")
	log.Println(largestChannelCapacitySats)
	log.Println("Estimated days spent earnings:")
	log.Println(secondsEarning / 86400.00)
	log.Println("Projected % APY:")
	log.Println(projectedAPY)
	log.Println("NOTE THIS EXCLUDEDS ON-CHAIN FEES FOR OPENING CHANNELS, GAINS FROM STRIKE, ROUTING FEES EARNED VIA DEEZY, AND ROUTING FEES PAID ON LN")

	return satsEarned, err
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
