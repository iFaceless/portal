package portal

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

var (
	ErrUnmatchedBrackets  = errors.New("unmatched brackets")
	ErrPrefixIsNotBracket = errors.New("filter string must starts with '['")
)

var (
	cachedParseResultMap = make(map[string]map[int][]*FilterNode)
	lock                 sync.RWMutex
)

type FilterNode struct {
	Name     string        `json:"name"`
	Parent   *FilterNode   `json:"-"`
	Children []*FilterNode `json:"children"`
}

func (node *FilterNode) String() string {
	data, _ := json.MarshalIndent(node, "", "  ")
	return string(data)
}

type ExtractOption struct {
	ignoreNodeWithChildren bool
	queryByParentName      string
}

func ExtractFilterNodeNames(nodes []*FilterNode, opt *ExtractOption) []string {
	if len(nodes) == 0 {
		return nil
	}

	if opt == nil {
		opt = &ExtractOption{}
	}

	names := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if opt.queryByParentName != "" && n.Parent.Name != opt.queryByParentName {
			continue
		}

		if opt.ignoreNodeWithChildren && len(n.Children) > 0 {
			continue
		}

		names = append(names, n.Name)
	}
	return names
}

func ParseFilters(filters []string) (map[int][]*FilterNode, error) {
	return ParseFilterString("[" + strings.Join(filters, ",") + "]")
}

// ParseFilterString parses filter string to a filter tree (with extra levels).
// Example input:
// 1. [speaker[id,name]]
// 2. [speaker[id,name,vip_info[type,is_active]]]
func ParseFilterString(s string) (map[int][]*FilterNode, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	if !strings.HasPrefix(s, "[") {
		return nil, ErrPrefixIsNotBracket
	}

	lock.RLock()
	cachedResult, ok := cachedParseResultMap[s]
	if ok {
		lock.RUnlock()
		return cachedResult, nil
	}
	lock.RUnlock()

	lock.Lock()
	// don't care about non-ascii chars.
	filterInBytes := []byte(s)
	err := checkBracketPairs(filterInBytes)
	if err != nil {
		lock.Unlock()
		return nil, errors.WithStack(err)
	}

	result := doParse(filterInBytes)
	cachedParseResultMap[s] = result
	lock.Unlock()

	return result, nil
}

func checkBracketPairs(s []byte) error {
	stack := NewStack()
	for _, c := range s {
		switch c {
		case '[':
			stack.Push(c)
		case ']':
			x, err := stack.Pop()
			if err != nil {
				return ErrUnmatchedBrackets
			}
			if xc, ok := x.(byte); ok {
				if xc == '[' {
					continue
				}
			}
			return ErrUnmatchedBrackets
		default:
			continue
		}
	}

	if stack.Size() > 0 {
		return ErrUnmatchedBrackets
	}

	return nil
}

// doParse parses filter string to a filter tree.
// Note: stupid & ugly~
func doParse(s []byte) map[int][]*FilterNode {
	var (
		wordBuf            []byte
		levelNodesMap      = make(map[int][]*FilterNode)
		levelParentNodeMap = map[int]*FilterNode{-1: nil}
		level              = -1
	)

	appendNodes := func() *FilterNode {
		if len(wordBuf) == 0 {
			return nil
		}

		nthLevelNodes, ok := levelNodesMap[level]
		if !ok || nthLevelNodes == nil {
			nthLevelNodes = make([]*FilterNode, 0)
		}

		node := &FilterNode{Name: string(wordBuf), Parent: levelParentNodeMap[level]}
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

	// scan the map again to build a filter tree
	for i := 0; i < len(levelNodesMap)-1; i++ {
		for _, parentNode := range levelNodesMap[i] {
			for _, childNode := range levelNodesMap[i+1] {
				if childNode.Parent == parentNode {
					parentNode.Children = append(parentNode.Children, childNode)
				}
			}
		}
	}

	return levelNodesMap
}
