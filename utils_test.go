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
	Kind MessageKind
}

func (c *Car) Name() string {
	return c.name
}

func (c *Car) Country() Country {
	return Country{"China"}
}

type BigCar struct {
	Car
}

func (bc *BigCar) Name() string {
	return "big car"
}

func (bc *BigCar) Kind() MessageKind {
	return MessageKind(2)
}

func TestGetNestedValue_Ok(t *testing.T) {
	ctx := context.TODO()

	c := Car{name: "xixi", Kind: MessageKind(1)}
	r, e := nestedValue(ctx, c, []string{"Name"})
	assert.Nil(t, e)
	assert.Equal(t, "xixi", r.(string))

	r, e = nestedValue(ctx, &c, []string{"Name"})
	assert.Nil(t, e)
	assert.Equal(t, "xixi", r.(string))

	r, e = nestedValue(ctx, &c, []string{"Country", "Name"})
	assert.Nil(t, e)
	assert.Equal(t, "China", r.(string))

	mk := MessageKind(1)
	r, e = nestedValue(ctx, &mk, []string{"Name"})
	assert.Nil(t, e)
	assert.Equal(t, "ok", r)

	r, e = nestedValue(ctx, &c, []string{"Kind", "Alias"})
	assert.Nil(t, e)
	assert.Equal(t, "alias_ok", r)

	bigCar := &BigCar{Car: c}
	r, e = nestedValue(ctx, bigCar, []string{"Name"})
	assert.Nil(t, e)
	assert.Equal(t, "big car", r)

	r, e = nestedValue(ctx, bigCar, []string{"Kind", "Alias"})
	assert.Nil(t, e)
	assert.Equal(t, "alias_failed", r)
}

func TestGetNestedValue_Error(t *testing.T) {
	ctx := context.TODO()

	_, e := nestedValue(ctx, nil, []string{"Name"})
	assert.EqualError(t, e, "object is nil")

	var c = Car{name: "foo", Kind: MessageKind(1)}
	_, e = nestedValue(ctx, &c, []string{"What"})
	assert.Nil(t, e)

	_, e = nestedValue(ctx, &c, []string{"Kind", "NotFound"})
	assert.Nil(t, e)
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
func TestInvokeStructMethod(t *testing.T) {
	book := Book{name: "Test"}
	ctx := context.TODO()
	ctx = context.WithValue(ctx, "key", "hello, world")

	ret, err := invokeMethodOfAnyType(ctx, book, "ShortName")
	assert.Nil(t, err)
	assert.Equal(t, "Test", ret)

	ret, err = invokeMethodOfAnyType(ctx, book, "FullName")
	assert.Nil(t, err)
	assert.Equal(t, "Prefix Test", ret)

	ret, err = invokeMethodOfAnyType(ctx, &book, "ShortName")
	assert.Nil(t, err)
	assert.Equal(t, "Test", ret)

	ret, err = invokeMethodOfAnyType(ctx, &book, "FullName")
	assert.Nil(t, err)
	assert.Equal(t, "Prefix Test", ret)

	ret, err = invokeMethodOfAnyType(ctx, book, "AddBook", "Book 2")
	assert.Nil(t, err)
	assert.Equal(t, "Add 'Book 2' ok", ret)

	ret, err = invokeMethodOfAnyType(ctx, book, "GetContextKey")
	assert.Nil(t, err)
	assert.Equal(t, "hello, world", ret)

	ret, err = invokeMethodOfAnyType(ctx, book, "Plus", 100)
	assert.Nil(t, err)
	assert.Equal(t, 200, ret)

	ret, err = invokeMethodOfAnyType(ctx, book, "NoError")
	assert.Nil(t, err)
	assert.Equal(t, "no error", ret)

	ret, err = invokeMethodOfAnyType(ctx, &book, "MethodNotFound")
	assert.Errorf(t, err, "method '*portal.Book.MethodNotFound' not found in 'Book'")

	_, err = invokeMethodOfAnyType(ctx, &book, "Plus")
	assert.NotNil(t, err)
	assert.Equal(t, "method '*portal.Book.Plus' must has minimum 2 params: 1", err.Error())

	_, err = invokeMethodOfAnyType(ctx, &book, "Plus", 1, 2)
	assert.NotNil(t, err)
	assert.Equal(t, "method '*portal.Book.Plus' must has 2 params: 3", err.Error())

	_, err = invokeMethodOfAnyType(ctx, &book, "NotReturn")
	assert.NotNil(t, err)
	assert.Equal(t, "method '*portal.Book.NotReturn' must returns one result with an optional error", err.Error())

	_, err = invokeMethodOfAnyType(ctx, &book, "ReturnTooMany")
	assert.NotNil(t, err)
	assert.Equal(t, "method '*portal.Book.ReturnTooMany' must returns one result with an optional error", err.Error())

	_, err = invokeMethodOfAnyType(ctx, &book, "ReturnError")
	assert.NotNil(t, err)
	assert.Equal(t, "error", err.Error())

	_, err = invokeMethodOfAnyType(ctx, &book, "LastReturnValueNotErrorType")
	assert.NotNil(t, err)
	assert.Equal(t, "the last return value of method '*portal.Book.LastReturnValueNotErrorType' must be of `error` type", err.Error())
}

type MessageKind int

func (mk *MessageKind) Name() string {
	switch *mk {
	case 1:
		return "ok"
	case 2:
		return "failed"
	default:
		return "undefined"
	}
}

func (mk MessageKind) Alias() string {
	switch mk {
	case 1:
		return "alias_ok"
	case 2:
		return "alias_failed"
	default:
		return "alias_undefined"
	}
}

func (mk MessageKind) ValueWithArgs(a string) string {
	switch mk {
	case 1:
		return "ok" + a
	case 2:
		return "failed" + a
	default:
		return "undefined" + a
	}
}

func TestInvokeMethodOfNonStruct(t *testing.T) {
	ctx := context.Background()
	mk := MessageKind(1)
	ret1, err1 := invokeMethodOfAnyType(ctx, mk, "Name")
	assert.Nil(t, err1)
	assert.Equal(t, "ok", ret1.(string))

	ret2, err2 := invokeMethodOfAnyType(ctx, &mk, "Name")
	assert.Nil(t, err2)
	assert.Equal(t, "ok", ret2.(string))

	ret3, err3 := invokeMethodOfAnyType(ctx, mk, "Alias")
	assert.Nil(t, err3)
	assert.Equal(t, "alias_ok", ret3.(string))

	ret4, err4 := invokeMethodOfAnyType(ctx, &mk, "Alias")
	assert.Nil(t, err4)
	assert.Equal(t, "alias_ok", ret4.(string))

	ret5, err5 := invokeMethodOfAnyType(ctx, &mk, "ValueWithArgs", "_hello")
	assert.Nil(t, err5)
	assert.Equal(t, "ok_hello", ret5.(string))

	_, err6 := invokeMethodOfAnyType(ctx, &mk, "NotFound", "_hello")
	assert.NotNil(t, err6)
	assert.Equal(t, "method 'NotFound' not found in '*portal.MessageKind'", err6.Error())
}

//1000000	      1371 ns/op
func BenchmarkInvokeMethod(b *testing.B) {
	b.ResetTimer()

	book := Book{name: "Test"}
	for i := 0; i < b.N; i++ {
		_, _ = invokeMethodOfAnyType(context.TODO(), book, "FullName")
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
