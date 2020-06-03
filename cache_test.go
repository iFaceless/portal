package portal

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	shortNameCounter int
	fullNameCounter  int
	infoCounter      int
	nameCounter      int
)

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
	Name     string           `json:"name" portal:"attr:Name"`
}

type StudentSchema struct {
	FullName  string `json:"full_name,omitempty" portal:"attr:FullName"`
	ShortName string `json:"short_name,omitempty" portal:"meth:GetShortName"`
	Age       int    `json:"age" portal:"attr:Info.Age;async"`
	Height    int    `json:"height" portal:"attr:Info.Height;async"`
}

func (sch *StudentSchema) GetShortName(s *Student) string {
	shortNameCounter += 1
	return string([]rune(s.FirstName)[0]) + string([]rune(s.LastName)[0])
}

func TestDumpWithCache(t *testing.T) {
	SetCache(DefaultCache)
	defer SetCache(nil)
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
}

func TestDumpNestedWithCache(t *testing.T) {
	SetCache(DefaultCache)
	defer SetCache(nil)
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
	err := Dump(&cc, c)
	assert.Nil(t, err)

	assert.Equal(t, 1, shortNameCounter)
	assert.Equal(t, 1, fullNameCounter)
	assert.Equal(t, 1, infoCounter)
	assert.Equal(t, 1, nameCounter)
}

var metaCounter int

type Food struct {
	ID int
}

type meta struct {
	Weight int
	Size   int
}

func (f *Food) Meta() meta {
	metaCounter += 1
	return meta{Weight: 10, Size: 20}
}

type FoodSchema struct {
	Weight string `portal:"attr:Meta.Weight;async"`
	Size   string `portal:"attr:Meta.Size;async"`
}

func (s *FoodSchema) PortalDisableCache() bool {
	return true
}

type FoodSchemaTwo struct {
	Weight string `portal:"attr:Meta.Weight;disablecache;async"`
	Size   string `portal:"attr:Meta.Size;async"`
}

type FoodSchemaThree struct {
	Weight string `portal:"attr:Meta.Weight;async"`
	Size   string `portal:"attr:Meta.Size;async"`
}

func TestDumpWithCacheDisabled(t *testing.T) {
	SetCache(DefaultCache)
	defer SetCache(nil)
	metaCounter = 0

	f := Food{
		ID: 1,
	}

	var ff FoodSchema
	Dump(&ff, &f)
	assert.Equal(t, 2, metaCounter)
	metaCounter = 0

	var ff2 FoodSchemaTwo
	Dump(&ff2, &f)
	assert.Equal(t, 2, metaCounter)
	metaCounter = 0

	var ff3 FoodSchemaThree
	Dump(&ff3, &f)
	assert.Equal(t, 1, metaCounter)
	metaCounter = 0

	Dump(&ff3, &f, DisableCache())
	assert.Equal(t, 2, metaCounter)
	metaCounter = 0

	SetCache(nil)
	Dump(&ff3, &f)
	assert.Equal(t, 2, metaCounter)
	metaCounter = 0
}
