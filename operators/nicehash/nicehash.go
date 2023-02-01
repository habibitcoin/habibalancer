package nicehash

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/habibitcoin/habibalancer/configs"
)

const (
	niceHashURL = "https://api2.nicehash.com"

	balanceEndpoint = "/main/api/v2/accounting/account2/BTC" // GET

	invoicesEndpoint = "/main/api/v2/accounting/depositAddresses?amount=SUB_AMOUNT&currency=BTC&walletType=LIGHTNING" // GET

	withdrawEndpoint = "/main/api/v2/accounting/withdrawal" // POST
)

type NicehashClient struct {
	Client                     *http.Client
	ApiKey                     string
	ApiSecret                  string
	NicehashWithdrawAddressKey string
	OrganizationId             string
	Context                    context.Context
}

// func NewClient
func NewClient(ctx context.Context) (client NicehashClient) {
	httpClient := &http.Client{}
	config := configs.GetConfig(ctx)
	client = NicehashClient{
		Client:                     httpClient,
		ApiKey:                     config.NicehashApiKey,
		ApiSecret:                  config.NicehashApiSecret,
		NicehashWithdrawAddressKey: config.NicehashWithdrawAddressKey,
		OrganizationId:             config.NicehashOrganizationId,
		Context:                    ctx,
	}

	return client
}

type WithdrawPayload struct {
	Currency            string `json:"currency"`
	Amount              string `json:"amount"`
	WithdrawalAddressId string `json:"withdrawalAddressId"`
}

type WithdrawResponse struct {
	Id string `json:"id"`
}

// Receives an amount defined in BTC, returns success.
func (client NicehashClient) Withdraw() (string, error) {
	nicehashWithdrawAmtXBTmin, _ := strconv.ParseFloat(configs.GetConfig(client.Context).NicehashWithdrawBtcMin, 64)

	nicehashBalanceStringXBT, err := client.GetBalance()
	if err != nil {
		log.Println("Error fetching nicehash balance")
		log.Println(err)
		return "", err
	}
	log.Println("nicehash balance XBT")
	log.Println(nicehashBalanceStringXBT)
	nicehashBalanceFloatXBT, _ := strconv.ParseFloat(nicehashBalanceStringXBT, 64)

	if nicehashBalanceFloatXBT > nicehashWithdrawAmtXBTmin {
		resp, err := client.sendPostRequest(withdrawEndpoint, &WithdrawPayload{
			Currency:            "BTC",
			Amount:              nicehashBalanceStringXBT,
			WithdrawalAddressId: client.NicehashWithdrawAddressKey,
		})
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		bodyString := string(bodyBytes)
		log.Println(bodyString)
		withdrawId := WithdrawResponse{}
		json.Unmarshal(bodyBytes, &withdrawId)
		if withdrawId.Id != "" {
			return withdrawId.Id, nil
		} else {
			log.Println("Possible error withdrawing from Nicehash")
			log.Println(err)
			return "", err
		}

	}
	log.Println("Balance too low for withdrawal " + nicehashBalanceStringXBT)
	return "", nil
}

type BalanceResponse struct {
	Available string `json:"available"`
}

func (client NicehashClient) GetBalance() (string, error) {
	resp, err := client.sendGetRequest(balanceEndpoint)
	if err != nil {
		return "", err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Println(bodyString)
	balance := BalanceResponse{}
	json.Unmarshal(bodyBytes, &balance)

	return balance.Available, nil
}

type InvoiceResponse struct {
	List []struct {
		Type struct {
			Code                string      `json:"code"`
			Description         interface{} `json:"description"`
			SupportedCurrencies interface{} `json:"supportedCurrencies"`
		} `json:"type"`
		Address  string `json:"address"`
		Currency string `json:"currency"`
	} `json:"list"`
}

// Receives an amount defined in BTC, returns an invoice.
func (client NicehashClient) GetAddress(amount string) (string, error) {
	invoicesEndpoint2 := strings.Replace(invoicesEndpoint, "SUB_AMOUNT", amount, 1)

	resp, err := client.sendGetRequest(invoicesEndpoint2)
	if err != nil {
		return "", err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	log.Println(bodyString)
	invoice := InvoiceResponse{}
	json.Unmarshal(bodyBytes, &invoice)

	return invoice.List[0].Address, nil
}

func (client NicehashClient) sendGetRequest(endpoint string) (*http.Response, error) {
	log.Println(endpoint)

	req, err := http.NewRequest("GET", niceHashURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	nonce, _ := uuid.NewUUID()
	reqID, _ := uuid.NewUUID()
	timeStamp := strconv.Itoa(int(time.Now().UnixMilli()))
	req.Header.Add("X-Time", timeStamp)
	req.Header.Add("X-Nonce", nonce.String())
	req.Header.Add("X-Organization-Id", client.OrganizationId)
	req.Header.Add("X-Request-Id", reqID.String())
	digest := getDigest(client.ApiSecret, client.ApiKey, timeStamp, nonce.String(), client.OrganizationId, req.Method, req.URL.Path, req.URL.Query().Encode(), "")
	req.Header.Add("X-Auth", client.ApiKey+":"+digest)

	resp, err := client.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (client NicehashClient) sendPostRequest(endpoint string, payload interface{}) (*http.Response, error) {
	jsonStr, err := json.Marshal(payload)
	if err != nil {
		log.Println("Error marshaling")
		log.Println(err)
	}
	log.Println(string(jsonStr))

	log.Println(endpoint)

	req, err := http.NewRequest("POST", niceHashURL+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	nonce, _ := uuid.NewUUID()
	reqID, _ := uuid.NewUUID()
	timeStamp := strconv.Itoa(int(time.Now().UnixMilli()))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-Time", timeStamp)
	req.Header.Add("X-Nonce", nonce.String())
	req.Header.Add("X-Organization-Id", client.OrganizationId)
	req.Header.Add("X-Request-Id", reqID.String())
	digest := getDigest(client.ApiSecret, client.ApiKey, timeStamp, nonce.String(), client.OrganizationId, req.Method, req.URL.Path, req.URL.Query().Encode(), string(jsonStr))
	req.Header.Add("X-Auth", client.ApiKey+":"+digest)

	resp, err := client.Client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return resp, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func getDigest(APISecret, APIKey string, time string, nonce, XOrganizationId, method, endpoint, query, body string) string {

	message := APIKey + "\x00" + time + "\x00" + nonce + "\x00" + "\x00" + XOrganizationId + "\x00" + "\x00" + method + "\x00" + endpoint + "\x00" + query
	if method == http.MethodPost {
		message = message + "\x00" + body
	}

	h := hmac.New(sha256.New, []byte(APISecret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
