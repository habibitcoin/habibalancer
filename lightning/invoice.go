package lightning

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type InvoiceRequest struct {
	cltv_expiry      string
	add_index        string
	creation_date    string
	private          bool
	value            string `json:"value"`
	expiry           string
	fallback_addr    string
	r_hash           byte
	memo             string `json:"memo"`
	receipt          byte
	amt_paid_msat    string
	payment_request  string
	description_hash byte
	settle_index     string
	settle_date      string
	settled          bool
	r_preimage       byte
	amt_paid_sat     string
}

type InvoiceResponse struct {
	RHash          byte   `json:"r_hash"`
	PaymentRequest string `json:"payment_request"`
	SettleDate     string `json:"settle_date"`
	AddIndex       string `json:"add_index"`
	State          string `json:"state"`
	AmtSats        string `json:"amt_paid_sat"`
	AmtMsats       string `json:"amt_paid_msat"`
	CltvExpiry     string `json:"cltvy_expiry"`
	Htlcs          []Htlc `json:"htlcs"`
	IsKeysend      bool   `json:"is_keysend"`
	PaymentAddr    string `json:"payment_addr"`
}

type Htlc struct {
	ChanId  string `json:"chan_id"`
	State   string `json:"state"`
	AmtMsat string `json:"amt_msat"`
}

type InvoiceListResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
}

func (client *LightningClient) CreateInvoice(amount string) (string, error) {
	log.Println(amount)

	resp, err := client.sendPostRequest("v1/invoices", `{"value":"`+amount+`","memo":"`+amount+`"}`)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, err
}

func (client *LightningClient) GetInvoices() (invoices InvoiceListResponse, err error) {
	// First see if invoice exists
	resp, err := client.sendGetRequest("v1/invoices")
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return invoices, err
	}
	validInvoices := InvoiceListResponse{}
	json.Unmarshal(bodyBytes, &validInvoices)

	return validInvoices, nil
}

func (client *LightningClient) GetInvoicePaid(invoice InvoiceResponse) (bool, error) {
	var (
		invoiceValid   = false
		invoicePending = false
		invoicePaid    = false
	)

	// First see if invoice exists
	resp, err := client.sendGetRequest("v1/invoices")
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	validInvoices := InvoiceListResponse{}
	json.Unmarshal(bodyBytes, &validInvoices)

	for _, validInvoice := range validInvoices.Invoices {
		if validInvoice.PaymentRequest == invoice.PaymentRequest {
			invoiceValid = true
			break
		}
	}

	resp, err = client.sendGetRequest("v1/invoices?pending_only=true")
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	pendingInvoices := InvoiceListResponse{}
	json.Unmarshal(bodyBytes, &pendingInvoices)

	for _, pendingInvoice := range pendingInvoices.Invoices {
		if pendingInvoice.PaymentRequest == invoice.PaymentRequest {
			invoicePending = true
			break
		}
	}

	if invoiceValid && !invoicePending {
		invoicePaid = true
	}

	invoicePaid = true

	return invoicePaid, err
}
