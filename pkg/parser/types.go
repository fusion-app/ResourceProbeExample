package parser

import (
	"bytes"
	"fmt"
	"strings"
)

type PatchCreatorSpec struct {
	Selectors  []JSONSelector
	// FirstProbe bool
}

type JSONSelector struct {
	JQSelector string
	PatchPath  string
	PrevValue  interface{}
	ValueType  ValueTypeName
}

type ProbeActionStatus struct {
	ActionID    string      `json:"action_id"`
	ActionName  string      `json:"action_name"`
	ResourceID  string      `json:"resource_id"`
	RefResource RefResource `json:"resource_instance_id,omitempty"`
	UpdateTime  string      `json:"action_state_time,omitempty"`
	State       ActionState `json:"action_state"`
}

type RefResource struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
	Icon      string `json:"icon,omitempty"`
	Desc      string `json:"description,omitempty"`
}

type ActionState string

const (
	ActionWaiting ActionState = "0"
	ActionRunning ActionState = "1"
	ActionEnd     ActionState = "2"
)

type PatchActionStatus struct {
	ActionID    string      `json:"actionID"`
	ActionName  string      `json:"actionName"`
	RefResource RefResource `json:"refResource,omitempty"`
	UpdateTime  string      `json:"updateTime"`
	State       ActionState `json:"state"`
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
