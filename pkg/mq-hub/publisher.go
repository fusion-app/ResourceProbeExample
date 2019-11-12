package mqhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type KafkaPubMsg struct {
	Records []PubMsgRecord `json:"records"`
}

type PubMsgRecord struct {
	Value MessageSpec `json:"value"`
}

func Pub(address string, topic string, msg MessageSpec) error {
	url := fmt.Sprintf("http://%s/topics/%s", address, topic)

	pubMsg := KafkaPubMsg{
		Records: []PubMsgRecord{
			{ Value: msg },
		},
	}
	msgBody, err := json.Marshal(pubMsg)
	log.Printf("MsgBody: %s", string(msgBody))
	if err != nil {
		log.Fatalf("Decode json string error: %+v", err.Error())
	}

	resp, err := http.Post(url, "application/vnd.kafka.json.v2+json", bytes.NewBuffer(msgBody))
	if err != nil {
		log.Fatalf("Pub msg to %s error: %+v", url, err.Error())
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Read Pub request error: %+v", err.Error())
	}
	log.Printf("Pub response: %v", string(respBody))

	return nil
}
