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
	AddIndex       string `json:"add_index"`
}

type InvoiceListResponse struct {
	Invoices []InvoiceResponse `json:"invoices"`
}

func CreateInvoice(amount string) (string, error) {
	log.Println(amount)

	resp, err := sendPostRequest("v1/invoices", `{"value":"`+amount+`","memo":"`+amount+`"}`)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	bodyString := string(bodyBytes)

	return bodyString, err
}

func GetInvoicePaid(invoice InvoiceResponse) (bool, error) {
	var (
		invoiceValid   = false
		invoicePending = false
		invoicePaid    = false
	)

	// First see if invoice exists
	resp, err := sendGetRequest("v1/invoices")
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

	resp, err = sendGetRequest("v1/invoices?pending_only=true")
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
