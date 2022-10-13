package lightning

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
)

type MessageResponse struct {
	Signature string `json:"signature"`
}

func (client *LightningClient) SignMessage(message string) (signature MessageResponse, err error) {
	messageUrl := base64.URLEncoding.EncodeToString([]byte(message))
	resp, err := client.sendPostRequest("v1/signmessage", `{"msg":"`+messageUrl+`"}`)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return signature, err
	}

	signature = MessageResponse{}
	json.Unmarshal(bodyBytes, &signature)

	return signature, err
}
