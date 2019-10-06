package portal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestField_String(t *testing.T) {
	type FooSchema struct {
		Name string
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Name"))
	assert.Equal(t, "FooSchema.Name", f.String())
}

func TestField_IsRequired(t *testing.T) {
	type FooSchema struct {
		Name string
		ID   string `portal:"required"`
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Name"))
	assert.False(t, f.IsRequired())

	f = NewField(schema, schema.Struct().Field("ID"))
	assert.True(t, f.IsRequired())
}

func TestField_IsNested(t *testing.T) {
	type BarSchema struct{}
	type FooSchema struct {
		Name string
		Bar  *BarSchema `portal:"nested"`
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Name"))
	assert.False(t, f.IsNested())

	f = NewField(schema, schema.Struct().Field("Bar"))
	assert.True(t, f.IsNested())
}

func TestField_Many(t *testing.T) {
	type BarSchema struct{}
	type FooSchema struct {
		Name string
		Bars []*BarSchema
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Name"))
	assert.False(t, f.Many())

	f = NewField(schema, schema.Struct().Field("Bars"))
	assert.True(t, f.Many())
}

func TestField_Method(t *testing.T) {
	type FooSchema struct {
		Name string
		Bar  string `portal:"meth:GetBar"`
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Bar"))
	assert.Equal(t, "GetBar", f.Method())
}

func TestField_HasMethod(t *testing.T) {
	type FooSchema struct {
		Name string
		Bar  string `portal:"meth:GetBar"`
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Bar"))
	assert.True(t, f.HasMethod())

	f = NewField(schema, schema.Struct().Field("Name"))
	assert.False(t, f.HasMethod())
}

func TestField_ChainingAttrs(t *testing.T) {
	type FooSchema struct {
		Name string
		Bazz string `portal:"attr:Bar.Bazz"`
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Bazz"))
	assert.Equal(t, []string{"Bar", "Bazz"}, f.ChainingAttrs())

	f = NewField(schema, schema.Struct().Field("Name"))
	assert.Equal(t, []string(nil), f.ChainingAttrs())
}

func TestField_HasChainingAttrs(t *testing.T) {
	type FooSchema struct {
		Name string
		Bazz string `portal:"attr:Bar.Bazz"`
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Bazz"))
	assert.True(t, f.HasChainingAttrs())

	f = NewField(schema, schema.Struct().Field("Name"))
	assert.False(t, f.HasChainingAttrs())
}

func TestField_ConstValue(t *testing.T) {
	type FooSchema struct {
		ID   int
		Name string `portal:"const:foo"`
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Name"))
	assert.Equal(t, "foo", f.Const())
	assert.True(t, f.HasConst())

	f = NewField(schema, schema.Struct().Field("ID"))
	assert.Equal(t, interface{}(nil), f.Const())
	assert.False(t, f.HasConst())
}

func TestField_Async(t *testing.T) {
	type FooSchema struct {
		ID   int
		Name string `portal:"async"`
	}

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Name"))
	assert.True(t, f.Async())

	f = NewField(schema, schema.Struct().Field("ID"))
	assert.False(t, f.Async())
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

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Bar"))
	assert.Equal(t, []string{"Name"}, f.NestedOnlyNames(nil))
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

	schema := NewSchema(&FooSchema{})
	f := NewField(schema, schema.Struct().Field("Bar"))
	assert.Equal(t, []string{"Name"}, f.NestedExcludeNames(nil))
}

type Person struct {
	Name string
}

func (p *Person) Value() (interface{}, error) {
	return p.Name, nil
}

type Timestamp int

func (t *Timestamp) SetValue(v interface{}) error {
	*t = Timestamp(v.(time.Time).Unix())
	return nil
}

func TestField_SetValue(t *testing.T) {
	type BarSchema struct {
		ID   string
		Name string
		Ts   *Timestamp
		Ts2  Timestamp
	}

	schema := NewSchema(&BarSchema{})
	f := NewField(schema, schema.Struct().Field("ID"))
	assert.Nil(t, f.SetValue(10))
	assert.Equal(t, "10", f.Value().(string))

	f = NewField(schema, schema.Struct().Field("Name"))
	assert.Nil(t, f.SetValue(&Person{Name: "foo"}))
	assert.Equal(t, "foo", f.Value().(string))

	f = NewField(schema, schema.Struct().Field("Ts"))

	now := time.Now()
	assert.Nil(t, f.SetValue(now))
	assert.Equal(t, Timestamp(now.Unix()), *f.Value().(*Timestamp))

	f = NewField(schema, schema.Struct().Field("Ts2"))
	assert.Nil(t, f.SetValue(now))
	assert.Equal(t, Timestamp(now.Unix()), f.Value().(Timestamp))
}

// BenchmarkNewField-4   	 3622506	       317 ns/op
func BenchmarkNewField(b *testing.B) {
	type FooSchema struct {
		Name string `portal:"async;meth:GetName"`
	}

	schema := NewSchema(&FooSchema{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewField(schema, schema.Struct().Field("Name"))
	}
}
