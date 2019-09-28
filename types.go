package portal

type Valuer interface {
	Value() (interface{}, error)
}

type ValueSetter interface {
	SetValue(v interface{}) error
}
