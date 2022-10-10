package lightning

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

//NESTED_WITNESS_PUBKEY_HASH

type AddressResponse struct {
	Address string `json:"address"`
}

func CreateAddress() (string, error) {
	resp, err := sendGetRequest("v1/newaddress?type=2")

	var address AddressResponse

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	json.Unmarshal(bodyBytes, &address)

	return address.Address, nil
}
