package deezy

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/habibitcoin/habibalancer/lightning"
)

// Closes a channel to Deezy.io when provided a channel point - returns response body as a string.
func CalculateEarnings(lightningClient lightning.LightningClient) (stats string, err error) {
	var (
		satsEarned        = 0
		satsPaidLooping   = 0
		satsPaidOnchain   = 0
		satsEarnedRouting = 0
		firstInvoice      = "0"
	)

	channelIDs, openingTxs, largestChannelCapacitySats, err := lightningClient.ListClosedChannels(lightningClient.DeezyPeer)
	if err != nil {
		log.Println("Error fetching closed channels for Deezy")
		return stats, err
	}

	invoicesReceived, err := lightningClient.GetInvoices()
	if err != nil {
		log.Println("Error fetching invoices from our node")
		return stats, err
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

	// calculate fees paid to perform looping operations
	completedPayments, err := lightningClient.ListPayments()
	if err != nil {
		return stats, err
	}

	for _, payment := range completedPayments.Payments {
		deezyPayment := false
		for _, htlc := range payment.Htlcs {
			for _, hop := range htlc.Route.Hops {
				if hop.CustomRecords.DeezyRecord == "Akv68Mq+f4dP0z6/fG9OU4WXH8UE7z9JJDLp4+x34bXP" {
					deezyPayment = true
					break
				}
			}
			if deezyPayment {
				break
			}
		}
		if deezyPayment {
			amount, _ := strconv.Atoi(payment.FeeSat)
			satsPaidLooping = satsPaidLooping + amount
		}
	}

	// calcualte fees paid to open channels to Deezy
	transactions, err := lightningClient.ListTransactions()
	if err != nil {
		log.Println("Error fetching transactions from our node")
		return stats, err
	}

	for _, transaction := range transactions.Transactions {
		if contains(openingTxs, transaction.TxHash) {
			amount, _ := strconv.Atoi(transaction.TotalFees)
			satsPaidOnchain = satsPaidOnchain + amount
		}
	}

	// calculate routing fees earned
	forwards, err := lightningClient.ListForwards()
	if err != nil {
		log.Println("Error fetching forwards from our node")
		return stats, err
	}

	for _, forward := range forwards.Forwards {
		if contains(channelIDs, forward.ChanIdOut) {
			amount, _ := strconv.Atoi(forward.Fee)
			satsEarnedRouting = satsEarnedRouting + amount
		}
	}

	profit := satsEarned + satsEarnedRouting - satsPaidLooping - satsPaidOnchain

	firstInvoiceInt, _ := strconv.Atoi(firstInvoice)
	secondsEarning := float64(int(time.Now().Unix()) - firstInvoiceInt)
	daysEarning := secondsEarning / 86400
	normalizeYearFactor := 365 / daysEarning
	normalizedEarnings := float64(profit) * normalizeYearFactor
	projectedAPY := (normalizedEarnings / float64(largestChannelCapacitySats)) * 100

	description := fmt.Sprintf(
		`Deezy has paid us %d sats to date!
		Routing fees earned via Deezy channels: %v
		Routing fees paid in sats to perform looping operations: %v
		Onchain fees paid in sats to open channels to Deezy: %v
		Total Profit or Loss: %v
		Estimated working capital in sats: %v
		Estimated days spent earning: %f
		Projected APY %%: %f
		NOTE THIS EXCLUDEDS EARNINGS FROM STRIKE `,
		satsEarned,
		satsEarnedRouting,
		satsPaidLooping,
		satsPaidOnchain,
		profit,
		largestChannelCapacitySats,
		secondsEarning/86400.00,
		projectedAPY,
	)
	log.Println(description)

	s := DeezyStats{
		SatsEarnedDeezy:   satsEarned,
		SatsEarnedRouting: satsEarnedRouting,
		SatsPaidLooping:   satsPaidLooping,
		SatsPaidOnchain:   satsPaidOnchain,
		Profit:            profit,
		WorkingCapital:    largestChannelCapacitySats,
		DaysEarning:       secondsEarning / 86400.00,
		ProjectedAPY:      projectedAPY,
		Description:       description,
	}

	sJSON, _ := json.Marshal(s)

	return string(sJSON), err
}

type DeezyStats struct {
	SatsEarnedDeezy   int
	SatsEarnedRouting int
	SatsPaidLooping   int
	SatsPaidOnchain   int
	Profit            int
	WorkingCapital    int
	DaysEarning       float64
	ProjectedAPY      float64

	Description string
}
