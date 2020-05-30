package portal

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Student struct {
	ID        int
	FirstName string
	LastName  string
}

func (s *Student) FullName() string {
	return fmt.Sprintf("%s %s", s.FirstName, s.LastName)
}

func (s *Student) CacheID() string {
	return fmt.Sprintf("%d", s.ID)
}

type StudentSchema struct {
	FullName  string `json:"full_name,omitempty" portal:"attr:FullName"`
	ShortName string `json:"short_name,omitempty" portal:"meth:GetShortName"`
}

func (sch *StudentSchema) GetShortName(s *Student) string {
	return string([]rune(s.FirstName)[0]) + string([]rune(s.LastName)[0])
}

func TestDumpWithCache(t *testing.T) {
	SetCache(DefaultCache)

	s := Student{
		ID:        1,
		FirstName: "Harry",
		LastName:  "Potter",
	}

	var ss StudentSchema
	err := Dump(&ss, &s)
	assert.Nil(t, err)

	data, _ := json.Marshal(ss)
	assert.Equal(t, `{"full_name":"Harry Potter","short_name":"HP"}`, string(data))
}
