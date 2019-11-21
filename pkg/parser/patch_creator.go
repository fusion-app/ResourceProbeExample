package parser

import (
	"encoding/json"
	"fmt"
	"github.com/fusion-app/prober/pkg/mq-hub"
	"log"
	"time"
)

func (creator *PatchCreatorSpec) CreatePatches(json []byte) []mqhub.PatchItem {
	var patches []mqhub.PatchItem
	for idx, item := range creator.Selectors {
		value, err := JQParse(json, item.JQSelector, item.ValueType)
		if err != nil || value == nil {
			log.Printf("Not found value in json error by jq '%s': %+v", item.JQSelector, err.Error())
			continue
		}
		log.Printf("Parse value in json by '%s': %v, initValue: %v", item.JQSelector, value, item.PrevValue)
		if creator.FirstProbe || item.PrevValue != value {
			creator.Selectors[idx].PrevValue = value
			patches = append(patches, mqhub.PatchItem{
				Op:    mqhub.Add,
				Path:  item.PatchPath,
				Value: value,
			})
		} else {
			log.Printf("Value('%s') not changed", item.JQSelector)
		}
	}
	return patches
}

func CreateAppInstanceStatusPatches(response []byte) ([]mqhub.PatchItem, error) {
	var probeStatus []ProbeActionStatus
	if err := json.Unmarshal(response, &probeStatus); err != nil {
		return nil, fmt.Errorf("Invalid app instance status probe response: %+v ", err)
	}
	var patchStatus []PatchActionStatus
	for _, item := range probeStatus {
		patchStatus = append(patchStatus, PatchActionStatus{
			ActionID:    item.ActionID,
			ActionName:  item.ActionName,
			RefResource: item.RefResource,
			State:       item.State,
		})
	}
	if len(patchStatus) > 0 {
		patches := []mqhub.PatchItem{
			{
				Op:    mqhub.Add,
				Path:  "/actionStatus",
				Value: patchStatus,
			},
			{
				Op:    mqhub.Add,
				Path:  "/updateTime",
				Value: time.Now().String(),
			},
		}
		return patches, nil
	} else {
		return nil, fmt.Errorf("Empty app instance status probe response ")
	}
}