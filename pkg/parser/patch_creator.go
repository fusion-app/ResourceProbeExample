package parser

import (
	"bytes"
	"fmt"
	"github.com/fusion-app/ResourceProbeExample/pkg/mq-hub"
	"log"
	"strings"
)

type PatchCreatorSpec struct {
	Selectors  []JSONSelector
	FirstProbe bool
}

type JSONSelector struct {
	JQSelector string
	PatchPath  string
	PrevValue  interface{}
	ValueType  ValueTypeName
}

func (creator *PatchCreatorSpec) String() string {
	var b bytes.Buffer
	for _, selector := range creator.Selectors {
		b.WriteString(fmt.Sprintf(" { selector: %s; path: %s; type: %s } ", selector.JQSelector, selector.PatchPath, selector.ValueType))
	}
	return b.String()
}

func (creator *PatchCreatorSpec) Set(value string) error {
	items := strings.Split(value, ";")
	if len(items) < 3 {
		return fmt.Errorf("JSONSelector is invalid")
	}
	selector := JSONSelector{
		JQSelector: strings.TrimSpace(items[0]),
		PatchPath: strings.TrimSpace(items[1]),
		ValueType: ValueTypeName(strings.TrimSpace(items[2])),
	}
	creator.Selectors = append(creator.Selectors, selector)
	return nil
}

func (creator *PatchCreatorSpec) CreatePatches(json []byte) []mqhub.PatchItem {
	var patches []mqhub.PatchItem
	for idx, item := range creator.Selectors {
		value, err := JQParse(json, item.JQSelector, item.ValueType)
		if err != nil || value == nil {
			log.Printf("Not found value in json error by jq '%s': %+v", item.JQSelector, err.Error())
			continue
		}
		log.Printf("Parse value in json by '%s': %v, initValue: %v", item.JQSelector, value, item.PrevValue)
		if creator.FirstProbe {
			creator.Selectors[idx].PrevValue = value
		} else if item.PrevValue != value {
			creator.Selectors[idx].PrevValue = value
			patches = append(patches, mqhub.PatchItem{
				Op:    mqhub.Add,
				Path:  item.PatchPath,
				Value: fmt.Sprintf("%v", value),
			})
		} else {
			log.Printf("Value('%s') not changed", item.JQSelector)
		}
	}
	return patches
}
