package portal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFilterString(t *testing.T) {
	asserter := assert.New(t)

	asserter.Nil(ParseFilterString(""))
	asserter.Equal(&FilterNode{
		Key: "User",
	}, ParseFilterString("User"))
	asserter.Equal(&FilterNode{
		Key:   "User",
		Value: []string{"ID", "Name", "Age"},
	}, ParseFilterString("User[ID,Name,Age]"))
	asserter.Equal(&FilterNode{
		Key:   "User",
		Value: []string{"ID", "Name"},
		Children: []*FilterNode{
			{
				Key:   "School",
				Value: []string{"Name"},
			},
		},
	}, ParseFilterString("User[ID,Name,School[Name]]"))
}
