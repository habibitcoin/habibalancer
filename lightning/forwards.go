package lightning

import (
	"encoding/json"
	"io/ioutil"
	"time"
)

type ListForwardsResponse struct {
	Forwards []ListForwardReponse `json:"forwarding_events"`
}

type ListForwardReponse struct {
	Fee          string `json:"fee"`
	PeerAliasOut string `json:"peer_alias_out"`
	ChanIdOut    string `json:"chan_id_out"`
}

type ListTransactionPayload struct {
	StartTime    uint64 `json:"start_time"`
	EndTime      int64  `json:"end_time"`
	IndexOffset  uint32 `json:"index_offset"`
	NumMaxEvents uint32 `json:"num_max_events"`
	// PeerAliasLookup bool   `json:"peer_alias_lookup"`
}

func (client *LightningClient) ListForwards() (forwards ListForwardsResponse, err error) {
	payload := ListTransactionPayload{
		StartTime:    0,
		EndTime:      time.Now().Unix(),
		IndexOffset:  0,
		NumMaxEvents: 2147483647,
		// PeerAliasLookup: true,
	}
	resp, err := client.sendPostRequestJSON("v1/switch", payload)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return forwards, err
	}
	//log.Println(string(bodyBytes))
	json.Unmarshal(bodyBytes, &forwards)

	return forwards, err
}
