package portal

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestField_String(t *testing.T) {
	type FooSchema struct {
		Name string
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Name"))
	assert.Equal(t, "FooSchema.Name", f.String())
}

func TestField_Alias(t *testing.T) {
	type FooSchema struct {
		Name string `json:"name"`
	}

	schema := newSchema(&FooSchema{}).withFieldAliasMapTagName("json")
	f := newField(schema, schema.innerStruct().Field("Name"))
	assert.Equal(t, "name", f.alias)
}

func TestField_IsRequired(t *testing.T) {
	type FooSchema struct {
		Name string
		ID   string `portal:"required"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Name"))
	assert.False(t, f.isRequired())

	f = newField(schema, schema.innerStruct().Field("ID"))
	assert.True(t, f.isRequired())
}

func TestField_IsNested(t *testing.T) {
	type BarSchema struct{}
	type FooSchema struct {
		Name string
		Bar  *BarSchema `portal:"nested"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Name"))
	assert.False(t, f.isNested())

	f = newField(schema, schema.innerStruct().Field("Bar"))
	assert.True(t, f.isNested())
}

func TestField_Many(t *testing.T) {
	type BarSchema struct{}
	type FooSchema struct {
		Name string
		Bars []*BarSchema
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Name"))
	assert.False(t, f.hasMany())

	f = newField(schema, schema.innerStruct().Field("Bars"))
	assert.True(t, f.hasMany())
}

func TestField_Method(t *testing.T) {
	type FooSchema struct {
		Name string
		Bar  string `portal:"meth:GetBar"`
		Foo  string `portal:"meth:GetFoo.Name"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Bar"))
	m, attrs := f.method()
	assert.Equal(t, "GetBar", m)
	assert.Equal(t, 0, len(attrs))

	f = newField(schema, schema.innerStruct().Field("Foo"))
	m, attrs = f.method()
	assert.Equal(t, "GetFoo", m)
	assert.Equal(t, []string{"Name"}, attrs)
}

func TestField_HasMethod(t *testing.T) {
	type FooSchema struct {
		Name string
		Bar  string `portal:"meth:GetBar"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Bar"))
	assert.True(t, f.hasMethod())

	f = newField(schema, schema.innerStruct().Field("Name"))
	assert.False(t, f.hasMethod())
}

func TestField_ChainingAttrs(t *testing.T) {
	type FooSchema struct {
		Name string
		Bazz string `portal:"attr:Bar.Bazz"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Bazz"))
	assert.Equal(t, []string{"Bar", "Bazz"}, f.chainingAttrs())

	f = newField(schema, schema.innerStruct().Field("Name"))
	assert.Equal(t, []string(nil), f.chainingAttrs())
}

func TestField_HasChainingAttrs(t *testing.T) {
	type FooSchema struct {
		Name string
		Bazz string `portal:"attr:Bar.Bazz"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Bazz"))
	assert.True(t, f.hasChainingAttrs())

	f = newField(schema, schema.innerStruct().Field("Name"))
	assert.False(t, f.hasChainingAttrs())
}

func TestField_ConstValue(t *testing.T) {
	type FooSchema struct {
		ID   int
		Name string `portal:"const:foo"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Name"))
	assert.Equal(t, "foo", f.constValue())
	assert.True(t, f.hasConstValue())

	f = newField(schema, schema.innerStruct().Field("ID"))
	assert.Equal(t, interface{}(nil), f.constValue())
	assert.False(t, f.hasConstValue())
}

func TestField_DefaultValue(t *testing.T) {
	type UserSchema struct {
		ID string
	}

	type FooSchema struct {
		ID       *string           `json:"id" portal:"default:100"`
		Settings map[string]string `json:"settings" portal:"default:AUTO_INIT"`
		User     *UserSchema       `json:"user" portal:"default:AUTO_INIT"`
		Users    []*UserSchema     `json:"users" portal:"default:AUTO_INIT"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("ID"))
	assert.Equal(t, "100", f.defaultValue())
	assert.True(t, f.hasDefaultValue())

	f = newField(schema, schema.innerStruct().Field("Settings"))
	assert.Equal(t, map[string]string{}, f.defaultValue())
	assert.True(t, f.hasDefaultValue())

	f = newField(schema, schema.innerStruct().Field("User"))
	assert.Equal(t, &UserSchema{}, f.defaultValue())
	assert.True(t, f.hasDefaultValue())

	f = newField(schema, schema.innerStruct().Field("Users"))
	assert.Equal(t, make([]*UserSchema, 0), f.defaultValue())
	assert.True(t, f.hasDefaultValue())
}

func TestField_Async(t *testing.T) {
	type FooSchema struct {
		ID   int
		Name string `portal:"async"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Name"))
	assert.True(t, f.async())

	f = newField(schema, schema.innerStruct().Field("ID"))
	assert.False(t, f.async())
}

func TestField_NestedOnlyNames(t *testing.T) {
	type BarSchema struct {
		ID   string
		Name string
	}
	type FooSchema struct {
		Name string
		Bar  *BarSchema `portal:"only:Name"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Bar"))
	assert.Equal(t, []string{"Name"}, f.nestedOnlyNames(nil))
}

func TestField_NestedExcludeNames(t *testing.T) {
	type BarSchema struct {
		ID   string
		Name string
	}
	type FooSchema struct {
		Name string
		Bar  *BarSchema `portal:"exclude:Name"`
	}

	schema := newSchema(&FooSchema{})
	f := newField(schema, schema.innerStruct().Field("Bar"))
	assert.Equal(t, []string{"Name"}, f.nestedExcludeNames(nil))
}

type Person struct {
	Name string
}

func (p *Person) Value() (interface{}, error) {
	return p.Name, nil
}

type Timestamp struct {
	tm time.Time
}

func (t *Timestamp) SetValue(v interface{}) error {
	switch timeValue := v.(type) {
	case time.Time:
		t.tm = timeValue
	case *time.Time:
		t.tm = *timeValue
	default:
		return fmt.Errorf("expect type `time.Time`, not `%T`", v)
	}
	return nil
}

func (t *Timestamp) Value() (interface{}, error) {
	return t.tm, nil
}

type ErrorValue int

func (v ErrorValue) Value() (interface{}, error) {
	return 0, errors.New("error value")
}

func TestField_SetValue(t *testing.T) {
	type BarSchema struct {
		ID   string
		Name string
		Ts   *Timestamp
		Ts2  Timestamp
	}

	schema := newSchema(&BarSchema{})
	f := newField(schema, schema.innerStruct().Field("ID"))
	assert.Nil(t, f.setValue(10))
	assert.Equal(t, "10", f.Value().(string))

	f = newField(schema, schema.innerStruct().Field("Name"))
	assert.Nil(t, f.setValue(&Person{Name: "foo"}))
	assert.Equal(t, "foo", f.Value().(string))

	f = newField(schema, schema.innerStruct().Field("Ts"))
	now := time.Now()
	assert.Nil(t, f.setValue(now))
	assert.Equal(t, Timestamp{now}, *f.Value().(*Timestamp))

	f = newField(schema, schema.innerStruct().Field("Ts"))
	assert.Nil(t, f.setValue(&Timestamp{now}))
	assert.Equal(t, Timestamp{now}, *f.Value().(*Timestamp))

	f = newField(schema, schema.innerStruct().Field("Ts"))
	assert.Nil(t, f.setValue(Timestamp{now}))
	assert.Equal(t, Timestamp{now}, *f.Value().(*Timestamp))

	f = newField(schema, schema.innerStruct().Field("Ts2"))
	assert.Nil(t, f.setValue(&Timestamp{now}))
	assert.Equal(t, Timestamp{now}, f.Value().(Timestamp))

	f = newField(schema, schema.innerStruct().Field("Ts2"))
	assert.Nil(t, f.setValue(&Timestamp{now}))
	assert.Equal(t, Timestamp{now}, f.Value().(Timestamp))

	f = newField(schema, schema.innerStruct().Field("Ts2"))
	assert.Nil(t, f.setValue(now))
	assert.Equal(t, Timestamp{now}, f.Value().(Timestamp))

	f = newField(schema, schema.innerStruct().Field("Ts2"))
	assert.NotNil(t, f.setValue(100))

	f = newField(schema, schema.innerStruct().Field("Ts2"))
	assert.NotNil(t, f.setValue(ErrorValue(0)))
}

// BenchmarkNewField-4   	 3622506	       317 ns/op
func BenchmarkNewField(b *testing.B) {
	type FooSchema struct {
		Name string `portal:"async;meth:GetName"`
	}

	schema := newSchema(&FooSchema{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newField(schema, schema.innerStruct().Field("Name"))
	}
}

func Test_parseAlias(t *testing.T) {
	assert.Equal(t, "", parseAlias(""))
	assert.Equal(t, "", parseAlias("-"))
	assert.Equal(t, "id", parseAlias("id,empty"))
	assert.Equal(t, "urlToken", parseAlias("urlToken,empty"))
	assert.Equal(t, "url_token", parseAlias("url_token,empty"))
	assert.Equal(t, "url-token", parseAlias(" url-token, empty "))
}

// Benchmark_parseAlias-4   	35453392	        34.5 ns/op
func Benchmark_parseAlias(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseAlias(" url-token, empty ")
	}
}
