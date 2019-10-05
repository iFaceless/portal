package portal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFilterNodeNames(t *testing.T) {
	asserter := assert.New(t)

	nodeA := &FilterNode{Name: "A"}

	nodeB := &FilterNode{Name: "B"}
	nodeC := &FilterNode{Name: "C", Parent: nodeB}
	nodeD := &FilterNode{Name: "D", Parent: nodeB}
	nodeB.Children = append(nodeB.Children, nodeC, nodeD)

	nodeE := &FilterNode{Name: "E"}
	nodeF := &FilterNode{Name: "F", Parent: nodeE}
	nodeE.Children = append(nodeE.Children, nodeF)

	nodeG := &FilterNode{Name: "G"}
	nodesMap := map[int][]*FilterNode{
		0: {nodeA, nodeB, nodeE, nodeG},
		1: {nodeC, nodeD, nodeF},
	}

	names := ExtractFilterNodeNames(nodesMap[0], &ExtractOption{ignoreNodeWithChildren: true})
	asserter.Equal([]string{"A", "G"}, names)

	names = ExtractFilterNodeNames(nodesMap[0], &ExtractOption{ignoreNodeWithChildren: false})
	asserter.Equal([]string{"A", "B", "E", "G"}, names)

	names = ExtractFilterNodeNames(nodesMap[1], &ExtractOption{ignoreNodeWithChildren: false, queryByParentName: "B"})
	asserter.Equal([]string{"C", "D"}, names)
}

func TestParseFilters(t *testing.T) {
	asserter := assert.New(t)

	node, err := ParseFilters([]string{"A"})
	asserter.Nil(err)
	expected := map[int][]*FilterNode{
		0: {&FilterNode{Name: "A"}},
	}
	asserter.Equal(expected, node)

	node, err = ParseFilters([]string{"A", "B", "C"})
	asserter.Nil(err)
	expected = map[int][]*FilterNode{
		0: {
			&FilterNode{Name: "A"},
			&FilterNode{Name: "B"},
			&FilterNode{Name: "C"},
		},
	}
	asserter.Equal(expected, node)

	node, err = ParseFilters([]string{"A", "B[C,D]", "E[ F]", "G"})
	asserter.Nil(err)

	nodeA := &FilterNode{Name: "A"}

	nodeB := &FilterNode{Name: "B"}
	nodeC := &FilterNode{Name: "C", Parent: nodeB}
	nodeD := &FilterNode{Name: "D", Parent: nodeB}
	nodeB.Children = append(nodeB.Children, nodeC, nodeD)

	nodeE := &FilterNode{Name: "E"}
	nodeF := &FilterNode{Name: "F", Parent: nodeE}
	nodeE.Children = append(nodeE.Children, nodeF)

	nodeG := &FilterNode{Name: "G"}
	expected = map[int][]*FilterNode{
		0: {nodeA, nodeB, nodeE, nodeG},
		1: {nodeC, nodeD, nodeF},
	}
	asserter.Equal(expected, node)
}

func TestParseFilterString(t *testing.T) {
	asserter := assert.New(t)

	node, err := ParseFilterString("[A]")
	asserter.Nil(err)
	expected := map[int][]*FilterNode{
		0: {&FilterNode{Name: "A"}},
	}
	asserter.Equal(expected, node)

	node, err = ParseFilterString("[A,B, C ]")
	asserter.Nil(err)
	expected = map[int][]*FilterNode{
		0: {
			&FilterNode{Name: "A"},
			&FilterNode{Name: "B"},
			&FilterNode{Name: "C"},
		},
	}
	asserter.Equal(expected, node)

	node, err = ParseFilterString("[A,B[ C,D], E[F ],G ]")
	asserter.Nil(err)

	nodeA := &FilterNode{Name: "A"}

	nodeB := &FilterNode{Name: "B"}
	nodeC := &FilterNode{Name: "C", Parent: nodeB}
	nodeD := &FilterNode{Name: "D", Parent: nodeB}
	nodeB.Children = append(nodeB.Children, nodeC, nodeD)

	nodeE := &FilterNode{Name: "E"}
	nodeF := &FilterNode{Name: "F", Parent: nodeE}
	nodeE.Children = append(nodeE.Children, nodeF)

	nodeG := &FilterNode{Name: "G"}
	expected = map[int][]*FilterNode{
		0: {nodeA, nodeB, nodeE, nodeG},
		1: {nodeC, nodeD, nodeF},
	}
	asserter.Equal(expected, node)
}

func TestParseFilterString_BoundaryConditions(t *testing.T) {
	asserter := assert.New(t)
	node, err := ParseFilterString("")
	asserter.Nil(err)
	asserter.Nil(node)

	_, err = ParseFilterString("A")
	asserter.Equal(ErrPrefixIsNotBracket, err)
}

func Test_checkBracketPair(t *testing.T) {
	asserter := assert.New(t)
	asserter.Nil(checkBracketPairs([]byte("speaker")))
	asserter.Nil(checkBracketPairs([]byte("speaker[]")))
	asserter.Nil(checkBracketPairs([]byte("speaker[name,age[user[id]]]")))
	asserter.Equal(ErrUnmatchedBrackets, checkBracketPairs([]byte("speaker[")))
	asserter.Equal(ErrUnmatchedBrackets, checkBracketPairs([]byte("speaker]")))
	asserter.Equal(ErrUnmatchedBrackets, checkBracketPairs([]byte("speaker[user[id]]]")))
}
