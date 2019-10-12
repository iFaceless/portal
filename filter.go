package portal

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var (
	errUnmatchedBrackets  = errors.New("unmatched brackets")
	errPrefixIsNotBracket = errors.New("filter string must starts with '['")
)

var (
	cachedFilterResultMap sync.Map
)

type filterNode struct {
	Name     string        `json:"name"`
	Parent   *filterNode   `json:"-"`
	Children []*filterNode `json:"children"`
}

func (node *filterNode) String() string {
	data, _ := json.MarshalIndent(node, "", "  ")
	return string(data)
}

type extractOption struct {
	ignoreNodeWithChildren bool
	queryByParentName      string
	queryByParentNameAlias string
}

func extractFilterNodeNames(nodes []*filterNode, opt *extractOption) []string {
	if len(nodes) == 0 {
		return nil
	}

	if opt == nil {
		opt = &extractOption{}
	}

	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		var parentName string
		if n.Parent != nil {
			parentName = n.Parent.Name
		}

		if opt.queryByParentName != "" && parentName != opt.queryByParentName && parentName != opt.queryByParentNameAlias {
			continue
		}

		if opt.ignoreNodeWithChildren && len(n.Children) > 0 {
			continue
		}

		names = append(names, n.Name)
	}
	return names
}

func parseFilters(filters []string) (map[int][]*filterNode, error) {
	return parseFilterString("[" + strings.Join(filters, ",") + "]")
}

// parseFilterString parses filter string to a filter tree (with extra levels).
// Example input:
// 1. [speaker[id,name]]
// 2. [speaker[id,name,vip_info[type,is_active]]]
func parseFilterString(s string) (map[int][]*filterNode, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	if !strings.HasPrefix(s, "[") {
		return nil, errPrefixIsNotBracket
	}

	cachedResult, ok := cachedFilterResultMap.Load(s)
	if ok {
		rv, _ := cachedResult.(map[int][]*filterNode)
		return rv, nil
	}

	// don't care about non-ascii chars.
	filterInBytes := []byte(s)
	err := checkBracketPairs(filterInBytes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := doParse(filterInBytes)
	cachedFilterResultMap.Store(s, result)
	return result, nil
}

func checkBracketPairs(s []byte) error {
	stack := newStack()
	for _, c := range s {
		switch c {
		case '[':
			stack.push(c)
		case ']':
			x, err := stack.pop()
			if err != nil {
				return errUnmatchedBrackets
			}
			if xc, ok := x.(byte); ok {
				if xc == '[' {
					continue
				}
			}
			return errUnmatchedBrackets
		default:
			continue
		}
	}

	if stack.size() > 0 {
		return errUnmatchedBrackets
	}

	return nil
}

// doParse parses filter string to a filter tree.
// Note: stupid & ugly~
func doParse(s []byte) map[int][]*filterNode {
	var (
		wordBuf            []byte
		levelNodesMap      = make(map[int][]*filterNode)
		levelParentNodeMap = map[int]*filterNode{-1: nil}
		level              = -1
	)

	appendNodes := func() *filterNode {
		if len(wordBuf) == 0 {
			return nil
		}

		nthLevelNodes, ok := levelNodesMap[level]
		if !ok || nthLevelNodes == nil {
			nthLevelNodes = make([]*filterNode, 0)
		}

		node := &filterNode{Name: string(wordBuf), Parent: levelParentNodeMap[level]}
		if node.Parent != nil {
			node.Parent.Children = append(node.Parent.Children, node)
		}
		nthLevelNodes = append(nthLevelNodes, node)
		levelNodesMap[level] = nthLevelNodes
		wordBuf = make([]byte, 0)
		return node
	}

	for _, char := range s {
		switch char {
		case '\t', '\n', ' ':
			continue
		case ',':
			_ = appendNodes()
		case '[':
			node := appendNodes()
			level++
			if node != nil {
				levelParentNodeMap[level] = node
			}
		case ']':
			appendNodes()
			level--
		default:
			wordBuf = append(wordBuf, char)
		}
	}

	appendNodes()

	return levelNodesMap
}
