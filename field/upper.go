package field

import (
	"encoding/json"
	"strings"
)

type UpperString string

func (s *UpperString) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToUpper(string(*s)))
}
