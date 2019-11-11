package mqhub

import "time"

type MessageSpec struct {
	Target      TargetCRDSpec
	UpdatePatch []byte
	ProbeTime   time.Time
}

type TargetCRDSpec struct {
	UID       string
	Kind      string
	Name 	  string
	Namespace string
}
