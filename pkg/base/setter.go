package base

type Setter interface {
	Parse(src []byte) (interface{}, error)
	SetValue(interface{}) error
}
