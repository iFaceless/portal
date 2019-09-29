package portal

import (
	"context"
	"fmt"
	"testing"

	"reflect"

	"github.com/stretchr/testify/assert"
)

type Country struct{ Name string }

type Car struct {
	name string
}

func (c *Car) Name() string {
	return c.name
}

func (c *Car) Country() Country {
	return Country{"China"}
}

func TestGetNestedValue_Ok(t *testing.T) {
	ctx := context.TODO()

	c := Car{"xixi"}
	r, e := GetNestedValue(ctx, c, []string{"Name"})
	assert.Nil(t, e)
	assert.Equal(t, "xixi", r.(string))

	r, e = GetNestedValue(ctx, &c, []string{"Name"})
	assert.Nil(t, e)
	assert.Equal(t, "xixi", r.(string))

	r, e = GetNestedValue(ctx, &c, []string{"Country", "Name"})
	assert.Nil(t, e)
	assert.Equal(t, "China", r.(string))
}

func TestGetNestedValue_Error(t *testing.T) {
	ctx := context.TODO()

	_, e := GetNestedValue(ctx, nil, []string{"Name"})
	assert.EqualError(t, e, "object is nil")

	var m = 1
	_, e = GetNestedValue(ctx, &m, []string{"Name"})
	assert.EqualError(t, e, "object must be a struct or a pointer to struct")

	var c = Car{"foo"}
	_, e = GetNestedValue(ctx, &c, []string{"What"})
	assert.EqualError(t, e, "method 'What' not found in 'Car'")
}

type Foo struct{}

func TestStructName(t *testing.T) {
	asserter := assert.New(t)
	asserter.Equal("Foo", StructName(Foo{}))
	asserter.Equal("Foo", StructName(&Foo{}))
	asserter.PanicsWithValue("invalid struct type", func() {
		StructName("12")
	})
}

func TestAreIdenticalType(t *testing.T) {
	asserter := assert.New(t)

	foo1 := &Foo{}
	foo2 := &Foo{}
	asserter.True(Convertible(foo1, foo2))
	asserter.False(Convertible(foo1, Foo{}))
}

func TestIsNil(t *testing.T) {
	asserter := assert.New(t)
	asserter.True(IsNil(nil))

	foo := (*Foo)(nil)
	asserter.True(IsNil(foo))
}

type Book struct {
	name string
}

func (b Book) ShortName() string {
	return b.name
}

func (b *Book) FullName() string {
	return "Prefix " + b.name
}

func (b *Book) AddBook(name string) string {
	return fmt.Sprintf("Add '%s' ok", name)
}

func (b *Book) GetContextKey(ctx context.Context) string {
	return ctx.Value("key").(string)
}

func (b *Book) Plus(ctx context.Context, v int) int {
	return v + 100
}

//nolint
func TestInvokeMethod(t *testing.T) {
	book := Book{name: "Test"}
	ctx := context.TODO()
	ctx = context.WithValue(ctx, "key", "hello, world")

	ret, err := InvokeMethod(ctx, book, "ShortName")
	assert.Nil(t, err)
	assert.Equal(t, "Test", ret)

	ret, err = InvokeMethod(ctx, book, "FullName")
	assert.Nil(t, err)
	assert.Equal(t, "Prefix Test", ret)

	ret, err = InvokeMethod(ctx, &book, "ShortName")
	assert.Nil(t, err)
	assert.Equal(t, "Test", ret)

	ret, err = InvokeMethod(ctx, &book, "FullName")
	assert.Nil(t, err)
	assert.Equal(t, "Prefix Test", ret)

	ret, err = InvokeMethod(ctx, book, "AddBook", "Book 2")
	assert.Nil(t, err)
	assert.Equal(t, "Add 'Book 2' ok", ret)

	ret, err = InvokeMethod(ctx, book, "GetContextKey")
	assert.Nil(t, err)
	assert.Equal(t, "hello, world", ret)

	ret, err = InvokeMethod(ctx, book, "Plus", 100)
	assert.Nil(t, err)
	assert.Equal(t, 200, ret)

	ret, err = InvokeMethod(ctx, &book, "MethodNotFound")
	assert.Errorf(t, err, "method 'MethodNotFound' not found in 'Book'")
}

//1000000	      1371 ns/op
func BenchmarkInvokeMethod(b *testing.B) {
	b.ResetTimer()

	book := Book{name: "Test"}
	for i := 0; i < b.N; i++ {
		_, _ = InvokeMethod(context.TODO(), book, "FullName")
	}
}

func TestIndirectStructType(t *testing.T) {
	type Fruit struct{}

	fruitType := reflect.TypeOf(Fruit{})
	assert.Equal(t, fruitType, IndirectStructTypeP(reflect.TypeOf(Fruit{})))
	assert.Equal(t, fruitType, IndirectStructTypeP(reflect.TypeOf(&Fruit{})))
	assert.Equal(t, fruitType, IndirectStructTypeP(reflect.TypeOf([]Fruit{})))

	var fruits []*Fruit
	assert.Equal(t, fruitType, IndirectStructTypeP(reflect.TypeOf(fruits)))
	assert.Equal(t, fruitType, IndirectStructTypeP(reflect.TypeOf(&fruits)))

	var fruits2 []Fruit
	assert.Equal(t, fruitType, IndirectStructTypeP(reflect.TypeOf(fruits2)))
	assert.Equal(t, fruitType, IndirectStructTypeP(reflect.TypeOf(&fruits2)))
}

func TestMinInt(t *testing.T) {
	assert.Equal(t, 2, MinInt(2, 3))
	assert.Equal(t, 2, MinInt(3, 2))
	assert.Equal(t, 2, MinInt(2, 2))
}
