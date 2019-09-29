package portal

import (
	"strings"
)

func ParseFilters(fieldFilters []string) map[string]*FilterNode {
	nodes := make(map[string]*FilterNode, len(fieldFilters))
	for _, f := range fieldFilters {
		node := ParseFilterString(f)
		nodes[node.Key] = node
	}
	return nodes
}

type FilterNode struct {
	Key      string
	Value    []string
	Children []*FilterNode
}

// ParseFilterString parses filter string to a filter tree.
// Examples:
// - User -> {"User": nil}
// - User[ID,Name] -> {"User": ["ID", "Name"]}
// - User[ID,School[Name]] -> {"User": ["ID", "School": ["Name"]]}
func ParseFilterString(s string) *FilterNode {
	if s == "" {
		return nil
	}

	parts := strings.SplitN(s, "[", 2)
	if len(parts) == 1 {
		return &FilterNode{
			Key: parts[0],
		}
	}

	node := &FilterNode{
		Key: parts[0],
	}

	rightPart := parts[1]
	rightPart = strings.TrimSuffix(rightPart, "]")
	for _, field := range strings.Split(rightPart, ",") {
		if strings.Contains(field, "[") {
			if node.Children == nil {
				node.Children = make([]*FilterNode, 0)
			}
			node.Children = append(node.Children, ParseFilterString(field))
			continue
		}

		if node.Value == nil {
			node.Value = make([]string, 0)
		}

		node.Value = append(node.Value, field)
	}

	return node
}
