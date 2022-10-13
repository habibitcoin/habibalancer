package lightning

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type AddressResponse struct {
	Address string `json:"address"`
}

func (client *LightningClient) CreateAddress() (string, error) {
	resp, err := client.sendGetRequest("v1/newaddress?type=2")
	if err != nil {
		log.Println(err)
		return "", err
	}

	var address AddressResponse

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	if err := json.Unmarshal(bodyBytes, &address); err != nil {
		log.Println(err)
		return "", err
	}

	return address.Address, nil
}
