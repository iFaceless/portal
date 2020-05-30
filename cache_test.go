package portal

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var shortNameCounter int
var fullNameCounter int
var infoCounter int
var nameCounter int

type Class struct {
	Students []*Student
}

func (c *Class) Name() string {
	nameCounter += 1
	return "name"
}

type Student struct {
	ID        int
	FirstName string
	LastName  string
}

type studentInfo struct {
	Age    int
	Height int
}

func (s *Student) FullName() string {
	fullNameCounter += 1
	return fmt.Sprintf("%s %s", s.FirstName, s.LastName)
}

func (s *Student) Info() *studentInfo {
	time.Sleep(100 * time.Millisecond)
	infoCounter += 1
	return &studentInfo{
		Age:    17,
		Height: 177,
	}
}

type ClassSchema struct {
	Students []*StudentSchema `json:"students" portal:"nested;async"`
	Name     string           `json:"name" portal:"attr:Name;disablecache"`
}

type StudentSchema struct {
	FullName  string `json:"full_name,omitempty" portal:"attr:FullName"`
	ShortName string `json:"short_name,omitempty" portal:"meth:GetShortName"`
	Age       int    `json:"age" portal:"attr:Info.Age"`
	Height    int    `json:"height" portal:"attr:Info.Height"`
}

func (sch *StudentSchema) GetShortName(s *Student) string {
	shortNameCounter += 1
	return string([]rune(s.FirstName)[0]) + string([]rune(s.LastName)[0])
}

func TestDumpWithCache(t *testing.T) {
	SetCache(DefaultCache)
	defer DisableCache()
	shortNameCounter = 0
	fullNameCounter = 0
	infoCounter = 0

	s := Student{
		ID:        1,
		FirstName: "Harry",
		LastName:  "Potter",
	}

	var ss StudentSchema
	err := Dump(&ss, &s)
	assert.Nil(t, err)

	assert.Equal(t, 1, shortNameCounter)
	assert.Equal(t, 1, fullNameCounter)
	assert.Equal(t, 1, infoCounter)

	err = Dump(&ss, &s)
	assert.Nil(t, err)

	assert.Equal(t, 1, shortNameCounter)
	assert.Equal(t, 1, fullNameCounter)
	assert.Equal(t, 1, infoCounter)
}

func TestDumpNestedWithCache(t *testing.T) {
	SetCache(DefaultCache)
	defer DisableCache()
	shortNameCounter = 0
	fullNameCounter = 0
	infoCounter = 0
	nameCounter = 0

	c := Class{
		Students: []*Student{
			{
				ID:        1,
				FirstName: "Harry",
				LastName:  "Potter",
			},
		},
	}

	var cc ClassSchema
	err := Dump(&cc, &c)
	assert.Nil(t, err)

	assert.Equal(t, 1, shortNameCounter)
	assert.Equal(t, 1, fullNameCounter)
	assert.Equal(t, 1, infoCounter)
	assert.Equal(t, 1, nameCounter)

	err = Dump(&cc, &c)
	assert.Nil(t, err)

	assert.Equal(t, 1, shortNameCounter)
	assert.Equal(t, 1, fullNameCounter)
	assert.Equal(t, 1, infoCounter)
	// disablecache tag test
	assert.Equal(t, 2, nameCounter)
}
