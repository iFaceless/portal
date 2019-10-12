package portal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFilterNodeNames(t *testing.T) {
	asserter := assert.New(t)

	nodeA := &filterNode{Name: "A"}

	nodeB := &filterNode{Name: "B"}
	nodeC := &filterNode{Name: "C", Parent: nodeB}
	nodeD := &filterNode{Name: "D", Parent: nodeB}
	nodeB.Children = append(nodeB.Children, nodeC, nodeD)

	nodeE := &filterNode{Name: "E"}
	nodeF := &filterNode{Name: "F", Parent: nodeE}
	nodeE.Children = append(nodeE.Children, nodeF)

	nodeG := &filterNode{Name: "G"}

	nodeZ := &filterNode{Name: "z"}
	nodeH := &filterNode{Name: "h", Parent: nodeZ}
	nodeZ.Children = append(nodeZ.Children, nodeH)

	nodesMap := map[int][]*filterNode{
		0: {nodeA, nodeB, nodeE, nodeG, nodeZ},
		1: {nodeC, nodeD, nodeF, nodeH},
	}

	names := extractFilterNodeNames(nodesMap[0], &extractOption{ignoreNodeWithChildren: true})
	asserter.Equal([]string{"A", "G"}, names)

	names = extractFilterNodeNames(nodesMap[0], &extractOption{ignoreNodeWithChildren: false})
	asserter.Equal([]string{"A", "B", "E", "G", "z"}, names)

	names = extractFilterNodeNames(nodesMap[1], &extractOption{ignoreNodeWithChildren: false, queryByParentName: "B"})
	asserter.Equal([]string{"C", "D"}, names)

	names = extractFilterNodeNames(nodesMap[1], &extractOption{ignoreNodeWithChildren: false, queryByParentName: "Z", queryByParentNameAlias: "z"})
	asserter.Equal([]string{"h"}, names)
}

func TestParseFilters(t *testing.T) {
	asserter := assert.New(t)

	node, err := parseFilters([]string{"A"})
	asserter.Nil(err)
	expected := map[int][]*filterNode{
		0: {&filterNode{Name: "A"}},
	}
	asserter.Equal(expected, node)

	node, err = parseFilters([]string{"A", "B", "C"})
	asserter.Nil(err)
	expected = map[int][]*filterNode{
		0: {
			&filterNode{Name: "A"},
			&filterNode{Name: "B"},
			&filterNode{Name: "C"},
		},
	}
	asserter.Equal(expected, node)

	node, err = parseFilters([]string{"A", "B[C,D]", "E[ F]", "G"})
	asserter.Nil(err)

	nodeA := &filterNode{Name: "A"}

	nodeB := &filterNode{Name: "B"}
	nodeC := &filterNode{Name: "C", Parent: nodeB}
	nodeD := &filterNode{Name: "D", Parent: nodeB}
	nodeB.Children = append(nodeB.Children, nodeC, nodeD)

	nodeE := &filterNode{Name: "E"}
	nodeF := &filterNode{Name: "F", Parent: nodeE}
	nodeE.Children = append(nodeE.Children, nodeF)

	nodeG := &filterNode{Name: "G"}
	expected = map[int][]*filterNode{
		0: {nodeA, nodeB, nodeE, nodeG},
		1: {nodeC, nodeD, nodeF},
	}
	asserter.Equal(expected, node)
}

func TestParseFilterString(t *testing.T) {
	asserter := assert.New(t)

	node, err := parseFilterString("[A]")
	asserter.Nil(err)
	expected := map[int][]*filterNode{
		0: {&filterNode{Name: "A"}},
	}
	asserter.Equal(expected, node)

	node, err = parseFilterString("[A,B, C ]")
	asserter.Nil(err)
	expected = map[int][]*filterNode{
		0: {
			&filterNode{Name: "A"},
			&filterNode{Name: "B"},
			&filterNode{Name: "C"},
		},
	}
	asserter.Equal(expected, node)

	node, err = parseFilterString("[A,B[ C,D], E[F ],G ]")
	asserter.Nil(err)

	nodeA := &filterNode{Name: "A"}

	nodeB := &filterNode{Name: "B"}
	nodeC := &filterNode{Name: "C", Parent: nodeB}
	nodeD := &filterNode{Name: "D", Parent: nodeB}
	nodeB.Children = append(nodeB.Children, nodeC, nodeD)

	nodeE := &filterNode{Name: "E"}
	nodeF := &filterNode{Name: "F", Parent: nodeE}
	nodeE.Children = append(nodeE.Children, nodeF)

	nodeG := &filterNode{Name: "G"}
	expected = map[int][]*filterNode{
		0: {nodeA, nodeB, nodeE, nodeG},
		1: {nodeC, nodeD, nodeF},
	}
	asserter.Equal(expected, node)
}

func TestParseFilterString_BoundaryConditions(t *testing.T) {
	asserter := assert.New(t)
	node, err := parseFilterString("")
	asserter.Nil(err)
	asserter.Nil(node)

	_, err = parseFilterString("A")
	asserter.Equal(errPrefixIsNotBracket, err)
}

func Test_checkBracketPair(t *testing.T) {
	asserter := assert.New(t)
	asserter.Nil(checkBracketPairs([]byte("speaker")))
	asserter.Nil(checkBracketPairs([]byte("speaker[]")))
	asserter.Nil(checkBracketPairs([]byte("speaker[name,age[user[id]]]")))
	asserter.Equal(errUnmatchedBrackets, checkBracketPairs([]byte("speaker[")))
	asserter.Equal(errUnmatchedBrackets, checkBracketPairs([]byte("speaker]")))
	asserter.Equal(errUnmatchedBrackets, checkBracketPairs([]byte("speaker[user[id]]]")))
}
