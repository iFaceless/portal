package field

import (
	"fmt"
	"time"

	"github.com/spf13/cast"

	"github.com/pkg/errors"
)

type Timestamp time.Time

func (ts *Timestamp) Value() (interface{}, error) {
	return time.Time(*ts).Unix(), nil
}

func (ts *Timestamp) SetValue(v interface{}) error {
	switch tm := v.(type) {
	case time.Time:
		*ts = Timestamp(tm)
	case *time.Time:
		*ts = Timestamp(*tm)
	default:
		return errors.New("expect `time.Time` type")
	}
	return nil
}

func (ts *Timestamp) MarshalJSON() ([]byte, error) {
	v, err := ts.Value()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return []byte(fmt.Sprintf("%d", v.(int64))), nil
}

func (ts *Timestamp) UnmarshalJSON(v []byte) error {
	unixTime, err := cast.ToInt64E(string(v))
	if err != nil {
		return errors.WithStack(err)
	}

	*ts = Timestamp(time.Unix(unixTime, 0))
	return nil
}
