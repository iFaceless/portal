package field

import (
	"encoding/json"
	"strings"
)

type LowerString string

func (s *LowerString) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToLower(string(*s)))
}
