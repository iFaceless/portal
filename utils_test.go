package portal

import (
	"context"
	"fmt"
	"testing"

	"reflect"

	"github.com/pkg/errors"
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
	r, e := nestedValue(ctx, c, []string{"Name"})
	assert.Nil(t, e)
	assert.Equal(t, "xixi", r.(string))

	r, e = nestedValue(ctx, &c, []string{"Name"})
	assert.Nil(t, e)
	assert.Equal(t, "xixi", r.(string))

	r, e = nestedValue(ctx, &c, []string{"Country", "Name"})
	assert.Nil(t, e)
	assert.Equal(t, "China", r.(string))
}

func TestGetNestedValue_Error(t *testing.T) {
	ctx := context.TODO()

	_, e := nestedValue(ctx, nil, []string{"Name"})
	assert.EqualError(t, e, "object is nil")

	var m = 1
	_, e = nestedValue(ctx, &m, []string{"Name"})
	assert.EqualError(t, e, "object must be a struct or a pointer to struct")

	var c = Car{"foo"}
	_, e = nestedValue(ctx, &c, []string{"What"})
	assert.EqualError(t, e, "method 'What' not found in 'Car'")
}

type Foo struct{}

func TestStructName(t *testing.T) {
	asserter := assert.New(t)
	asserter.Equal("Foo", structName(Foo{}))
	asserter.Equal("Foo", structName(&Foo{}))
	asserter.PanicsWithValue("invalid struct type", func() {
		structName("12")
	})
}

func TestAreIdenticalType(t *testing.T) {
	asserter := assert.New(t)

	foo1 := &Foo{}
	foo2 := &Foo{}
	asserter.True(convertible(foo1, foo2))
	asserter.False(convertible(foo1, Foo{}))
}

func TestIsNil(t *testing.T) {
	asserter := assert.New(t)
	asserter.True(isNil(nil))

	foo := (*Foo)(nil)
	asserter.True(isNil(foo))
}

type Book struct {
	name string
}

func (b Book) NotReturn() {

}

func (b Book) ReturnTooMany() (string, string, error) {
	return "", "", nil
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

func (b *Book) NoError() (string, error) {
	return "no error", nil
}

func (b *Book) ReturnError() (string, error) {
	return "", errors.New("error")
}

func (b *Book) LastReturnValueNotErrorType() (string, string) {
	return "", ""
}

//nolint
func TestInvokeMethod(t *testing.T) {
	book := Book{name: "Test"}
	ctx := context.TODO()
	ctx = context.WithValue(ctx, "key", "hello, world")

	ret, err := invokeStructMethod(ctx, book, "ShortName")
	assert.Nil(t, err)
	assert.Equal(t, "Test", ret)

	ret, err = invokeStructMethod(ctx, book, "FullName")
	assert.Nil(t, err)
	assert.Equal(t, "Prefix Test", ret)

	ret, err = invokeStructMethod(ctx, &book, "ShortName")
	assert.Nil(t, err)
	assert.Equal(t, "Test", ret)

	ret, err = invokeStructMethod(ctx, &book, "FullName")
	assert.Nil(t, err)
	assert.Equal(t, "Prefix Test", ret)

	ret, err = invokeStructMethod(ctx, book, "AddBook", "Book 2")
	assert.Nil(t, err)
	assert.Equal(t, "Add 'Book 2' ok", ret)

	ret, err = invokeStructMethod(ctx, book, "GetContextKey")
	assert.Nil(t, err)
	assert.Equal(t, "hello, world", ret)

	ret, err = invokeStructMethod(ctx, book, "Plus", 100)
	assert.Nil(t, err)
	assert.Equal(t, 200, ret)

	ret, err = invokeStructMethod(ctx, book, "NoError")
	assert.Nil(t, err)
	assert.Equal(t, "no error", ret)

	ret, err = invokeStructMethod(ctx, &book, "MethodNotFound")
	assert.Errorf(t, err, "method 'MethodNotFound' not found in 'Book'")

	_, err = invokeStructMethod(ctx, &book, "Plus")
	assert.NotNil(t, err)
	assert.Equal(t, "method 'Plus' must has minimum 2 params: 1", err.Error())

	_, err = invokeStructMethod(ctx, &book, "Plus", 1, 2)
	assert.NotNil(t, err)
	assert.Equal(t, "method 'Plus' must has 2 params: 3", err.Error())

	_, err = invokeStructMethod(ctx, &book, "NotReturn")
	assert.NotNil(t, err)
	assert.Equal(t, "method 'NotReturn' must returns one result with an optional error", err.Error())

	_, err = invokeStructMethod(ctx, &book, "ReturnTooMany")
	assert.NotNil(t, err)
	assert.Equal(t, "method 'ReturnTooMany' must returns one result with an optional error", err.Error())

	_, err = invokeStructMethod(ctx, &book, "ReturnError")
	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())

	_, err = invokeStructMethod(ctx, &book, "LastReturnValueNotErrorType")
	assert.NotNil(t, err)
	assert.Equal(t, "the last return value of method 'LastReturnValueNotErrorType' must be of `error` type", err.Error())
}

//1000000	      1371 ns/op
func BenchmarkInvokeMethod(b *testing.B) {
	b.ResetTimer()

	book := Book{name: "Test"}
	for i := 0; i < b.N; i++ {
		_, _ = invokeStructMethod(context.TODO(), book, "FullName")
	}
}

func TestIndirectStructType(t *testing.T) {
	type Fruit struct{}

	fruitType := reflect.TypeOf(Fruit{})
	assert.Equal(t, fruitType, indirectStructTypeP(reflect.TypeOf(Fruit{})))
	assert.Equal(t, fruitType, indirectStructTypeP(reflect.TypeOf(&Fruit{})))
	assert.Equal(t, fruitType, indirectStructTypeP(reflect.TypeOf([]Fruit{})))

	var fruits []*Fruit
	assert.Equal(t, fruitType, indirectStructTypeP(reflect.TypeOf(fruits)))
	assert.Equal(t, fruitType, indirectStructTypeP(reflect.TypeOf(&fruits)))

	var fruits2 []Fruit
	assert.Equal(t, fruitType, indirectStructTypeP(reflect.TypeOf(fruits2)))
	assert.Equal(t, fruitType, indirectStructTypeP(reflect.TypeOf(&fruits2)))
}

func TestInnerStructType(t *testing.T) {
	type Foo struct {
		Name string
	}

	typ, _ := innerStructType(reflect.TypeOf(Foo{}))
	assert.Equal(t, reflect.TypeOf(Foo{}), typ)

	var foo Foo
	typ, _ = innerStructType(reflect.TypeOf(&foo))
	assert.Equal(t, reflect.TypeOf(Foo{}), typ)

	var foo2 *Foo
	typ, _ = innerStructType(reflect.TypeOf(&foo2))
	assert.Equal(t, reflect.TypeOf(Foo{}), typ)
}

func TestInnerStructTypeError(t *testing.T) {
	expectedError := errors.New("failed to get inner struct type")

	var a = 100
	_, err := innerStructType(reflect.TypeOf(a))
	assert.NotNil(t, err)
	assert.Equal(t, expectedError.Error(), err.Error())

	_, err = innerStructType(reflect.TypeOf(&a))
	assert.NotNil(t, err)
	assert.Equal(t, expectedError.Error(), err.Error())

	var b *int
	_, err = innerStructType(reflect.TypeOf(&b))
	assert.NotNil(t, err)
	assert.Equal(t, expectedError.Error(), err.Error())

	type Foo struct {
		Name string
	}
	var foo **Foo
	_, err = innerStructType(reflect.TypeOf(&foo))
	assert.NotNil(t, err)
	assert.Equal(t, "pointer level too deep", err.Error())
}

func BenchmarkInnerStructType(b *testing.B) {
	b.ResetTimer()

	type Foo struct {
		Name string
	}

	var bazz *Foo
	for i := 0; i < b.N; i++ {
		_, _ = innerStructType(reflect.TypeOf(&bazz))
	}
}
