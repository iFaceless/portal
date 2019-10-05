package portal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	intCases = []interface{}{
		100,
		int(100),
		int64(100),
		int32(100),
		int16(100),
		int8(100),
		uint(100),
		uint64(100),
		uint32(100),
		uint16(100),
		uint8(100),
		"100",
	}
)

type SuiteConvertTester struct {
	suite.Suite
}

func (s *SuiteConvertTester) TestToInt() {
	for _, c := range intCases {
		tmp := c

		var target int
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(100, out.(int))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(100, out.(int))
	}
}

func (s *SuiteConvertTester) TestToIntPtr() {
	for _, c := range intCases {
		tmp := c

		var target *int
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(100, *out.(*int))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(100, *out.(*int))
	}
}

func (s *SuiteConvertTester) TestToInt64() {
	for _, c := range intCases {
		tmp := c

		var target int64
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(int64(100), out.(int64))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int64(100), out.(int64))
	}
}

func (s *SuiteConvertTester) TestToInt64Ptr() {
	for _, c := range intCases {
		tmp := c

		var target *int64
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(int64(100), *out.(*int64))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int64(100), *out.(*int64))
	}
}

func (s *SuiteConvertTester) TestToInt32() {
	for _, c := range intCases {
		tmp := c

		var target int32
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(int32(100), out.(int32))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int32(100), out.(int32))
	}
}

func (s *SuiteConvertTester) TestToInt32Ptr() {
	for _, c := range intCases {
		tmp := c

		var target *int32
		out, err := Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int32(100), *out.(*int32))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int32(100), *out.(*int32))
	}
}

func (s *SuiteConvertTester) TestToInt16() {
	for _, c := range intCases {
		tmp := c

		var target int16
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(int16(100), out.(int16))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int16(100), out.(int16))
	}
}

func (s *SuiteConvertTester) TestToInt16Ptr() {
	for _, c := range intCases {
		tmp := c

		var target *int16
		out, err := Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int16(100), *out.(*int16))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int16(100), *out.(*int16))
	}
}

func (s *SuiteConvertTester) TestToInt8() {
	for _, c := range intCases {
		tmp := c

		var target int8
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(int8(100), out.(int8))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int8(100), out.(int8))
	}
}

func (s *SuiteConvertTester) TestToInt8Ptr() {
	for _, c := range intCases {
		tmp := c

		var target *int8
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(int8(100), *out.(*int8))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(int8(100), *out.(*int8))
	}
}

func (s *SuiteConvertTester) TestToUint() {
	for _, c := range intCases {
		tmp := c

		var target uint
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint(100), out.(uint))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint(100), out.(uint))
	}
}

func (s *SuiteConvertTester) TestToUintPtr() {
	for _, c := range intCases {
		tmp := c

		var target *uint
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint(100), *out.(*uint))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint(100), *out.(*uint))
	}
}

func (s *SuiteConvertTester) TestUioInt64() {
	for _, c := range intCases {
		tmp := c

		var target uint64
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint64(100), out.(uint64))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint64(100), out.(uint64))
	}
}

func (s *SuiteConvertTester) TestToUint64Ptr() {
	for _, c := range intCases {
		tmp := c

		var target *uint64
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint64(100), *out.(*uint64))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint64(100), *out.(*uint64))
	}
}

func (s *SuiteConvertTester) TestToUint32() {
	for _, c := range intCases {
		tmp := c

		var target uint32
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint32(100), out.(uint32))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint32(100), out.(uint32))
	}
}

func (s *SuiteConvertTester) TestToUint32Ptr() {
	for _, c := range intCases {
		tmp := c

		var target *uint32
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint32(100), *out.(*uint32))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint32(100), *out.(*uint32))
	}
}

func (s *SuiteConvertTester) TestToUint16() {
	for _, c := range intCases {
		tmp := c

		var target uint16
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint16(100), out.(uint16))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint16(100), out.(uint16))
	}
}

func (s *SuiteConvertTester) TestToUint16Ptr() {
	for _, c := range intCases {
		tmp := c

		var target *uint16
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint16(100), *out.(*uint16))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint16(100), *out.(*uint16))
	}
}

func (s *SuiteConvertTester) TestToUint8() {
	for _, c := range intCases {
		tmp := c

		var target uint8
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint8(100), out.(uint8))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint8(100), out.(uint8))
	}
}

func (s *SuiteConvertTester) TestToUint8Ptr() {
	for _, c := range intCases {
		tmp := c

		var target *uint8
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal(uint8(100), *out.(*uint8))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal(uint8(100), *out.(*uint8))
	}
}

func (s *SuiteConvertTester) TestToString() {
	for _, c := range intCases {
		tmp := c

		var target string
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal("100", out.(string))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal("100", out.(string))
	}
}

func (s *SuiteConvertTester) TestToStringPtr() {
	for _, c := range intCases {
		tmp := c

		var target *string
		out, err := Convert(target, tmp)
		s.Nil(err)
		s.Equal("100", *out.(*string))

		out, err = Convert(target, &tmp)
		s.Nil(err)
		s.Equal("100", *out.(*string))
	}
}

func (s *SuiteConvertTester) TestToTime() {
	var target time.Time

	t, _ := time.Parse("2006-01-02", "2019-10-05")
	out, err := Convert(target, "2019-10-05")
	s.Nil(err)
	s.Equal(t, out.(time.Time))

	x := "2019-10-05"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(t, out.(time.Time))
}

func (s *SuiteConvertTester) TestToTimePtr() {
	var target *time.Time

	t, _ := time.Parse("2006-01-02", "2019-10-05")
	out, err := Convert(target, "2019-10-05")
	s.Nil(err)
	s.Equal(t, *out.(*time.Time))

	x := "2019-10-05"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(t, *out.(*time.Time))
}

func (s *SuiteConvertTester) TestToDuration() {
	var target time.Duration

	d, _ := time.ParseDuration("300ms")

	out, err := Convert(target, "300ms")
	s.Nil(err)
	s.Equal(d, out.(time.Duration))

	x := "300ms"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(d, out.(time.Duration))
}

func (s *SuiteConvertTester) TestToDurationPtr() {
	var target *time.Duration

	d, _ := time.ParseDuration("300ms")

	out, err := Convert(target, "300ms")
	s.Nil(err)
	s.Equal(d, *out.(*time.Duration))

	x := "300ms"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(d, *out.(*time.Duration))
}

func (s *SuiteConvertTester) TestToBool() {
	var target bool

	out, err := Convert(target, "1")
	s.Nil(err)
	s.Equal(true, out.(bool))

	x := "0"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(false, out.(bool))
}

func (s *SuiteConvertTester) TestToBoolPtr() {
	var target *bool

	out, err := Convert(target, "1")
	s.Nil(err)
	s.Equal(true, *out.(*bool))

	x := "0"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(false, *out.(*bool))
}

func (s *SuiteConvertTester) TestToFloat32() {
	var target float32

	out, err := Convert(target, "1.234")
	s.Nil(err)
	s.Equal(float32(1.234), out.(float32))

	x := "1.234"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(float32(1.234), out.(float32))
}

func (s *SuiteConvertTester) TestToFloat32Ptr() {
	var target *float32

	out, err := Convert(target, "1.234")
	s.Nil(err)
	s.Equal(float32(1.234), *out.(*float32))

	x := "1.234"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(float32(1.234), *out.(*float32))
}

func (s *SuiteConvertTester) TestToFloat64() {
	var target float64

	out, err := Convert(target, "1.234")
	s.Nil(err)
	s.Equal(float64(1.234), out.(float64))

	x := "1.234"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(float64(1.234), out.(float64))
}

func (s *SuiteConvertTester) TestToFloat64Ptr() {
	var target *float64

	out, err := Convert(target, "1.234")
	s.Nil(err)
	s.Equal(float64(1.234), *out.(*float64))

	x := "1.234"
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(float64(1.234), *out.(*float64))
}

func (s *SuiteConvertTester) TestToStringMapString() {
	var target map[string]string

	x := map[interface{}]interface{}{
		"hello": "world",
	}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal(map[string]string{"hello": "world"}, out.(map[string]string))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(map[string]string{"hello": "world"}, out.(map[string]string))
}

func (s *SuiteConvertTester) TestToStringMapStringPtr() {
	var target *map[string]string

	x := map[interface{}]interface{}{
		"hello": "world",
	}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal(map[string]string{"hello": "world"}, *out.(*map[string]string))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(map[string]string{"hello": "world"}, *out.(*map[string]string))
}

func (s *SuiteConvertTester) TestToStringMapStringSlice() {
	var target map[string][]string

	x := map[interface{}]interface{}{
		"hello": []string{"world"},
	}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal(map[string][]string{"hello": {"world"}}, out.(map[string][]string))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(map[string][]string{"hello": {"world"}}, out.(map[string][]string))
}

func (s *SuiteConvertTester) TestToStringMapStringSlicePtr() {
	var target *map[string][]string

	x := map[interface{}]interface{}{
		"hello": []string{"world"},
	}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal(map[string][]string{"hello": {"world"}}, *out.(*map[string][]string))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(map[string][]string{"hello": {"world"}}, *out.(*map[string][]string))
}

func (s *SuiteConvertTester) TestToStringMapBool() {
	var target map[string]bool

	x := map[interface{}]interface{}{
		"hello": true,
	}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal(map[string]bool{"hello": true}, out.(map[string]bool))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(map[string]bool{"hello": true}, out.(map[string]bool))
}

func (s *SuiteConvertTester) TestToStringMapBoolPtr() {
	var target *map[string]bool

	x := map[interface{}]interface{}{
		"hello": true,
	}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal(map[string]bool{"hello": true}, *out.(*map[string]bool))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(map[string]bool{"hello": true}, *out.(*map[string]bool))
}

func (s *SuiteConvertTester) TestToStringMap() {
	var target map[string]interface{}

	x := map[interface{}]interface{}{
		"hello": true,
	}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal(map[string]interface{}{"hello": true}, out.(map[string]interface{}))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(map[string]interface{}{"hello": true}, out.(map[string]interface{}))
}

func (s *SuiteConvertTester) TestToStringMapPtr() {
	var target *map[string]interface{}

	x := map[interface{}]interface{}{
		"hello": true,
	}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal(map[string]interface{}{"hello": true}, *out.(*map[string]interface{}))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal(map[string]interface{}{"hello": true}, *out.(*map[string]interface{}))
}

func (s *SuiteConvertTester) TestToSlice() {
	var target []interface{}

	x := []interface{}{1, 2, 3}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal([]interface{}{1, 2, 3}, out.([]interface{}))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]interface{}{1, 2, 3}, out.([]interface{}))
}

func (s *SuiteConvertTester) TestToSlicePtr() {
	var target *[]interface{}

	x := []interface{}{1, 2, 3}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal([]interface{}{1, 2, 3}, *out.(*[]interface{}))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]interface{}{1, 2, 3}, *out.(*[]interface{}))
}

func (s *SuiteConvertTester) TestToBoolSlice() {
	var target []bool

	x := []interface{}{1, 0, "true", "false"}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal([]bool{true, false, true, false}, out.([]bool))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]bool{true, false, true, false}, out.([]bool))
}

func (s *SuiteConvertTester) TestToBoolSlicePtr() {
	var target *[]bool

	x := []interface{}{1, 0, "true", "false"}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal([]bool{true, false, true, false}, *out.(*[]bool))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]bool{true, false, true, false}, *out.(*[]bool))
}

func (s *SuiteConvertTester) TestToStringSlice() {
	var target []string

	x := []interface{}{1, "2", true}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal([]string{"1", "2", "true"}, out.([]string))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]string{"1", "2", "true"}, out.([]string))
}

func (s *SuiteConvertTester) TestToStringSlicePtr() {
	var target *[]string

	x := []interface{}{1, "2", true}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal([]string{"1", "2", "true"}, *out.(*[]string))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]string{"1", "2", "true"}, *out.(*[]string))
}

func (s *SuiteConvertTester) TestToIntSlice() {
	var target []int

	x := []interface{}{"1", "2"}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal([]int{1, 2}, out.([]int))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]int{1, 2}, out.([]int))
}

func (s *SuiteConvertTester) TestToIntSlicePtr() {
	var target *[]int

	x := []interface{}{"1", "2"}

	out, err := Convert(target, x)
	s.Nil(err)
	s.Equal([]int{1, 2}, *out.(*[]int))

	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]int{1, 2}, *out.(*[]int))
}

func (s *SuiteConvertTester) TestToDurationSlice() {
	var target []time.Duration

	d, _ := time.ParseDuration("300ms")

	out, err := Convert(target, []string{"300ms"})
	s.Nil(err)
	s.Equal([]time.Duration{d}, out.([]time.Duration))

	x := []string{"300ms"}
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]time.Duration{d}, out.([]time.Duration))
}

func (s *SuiteConvertTester) TestToDurationSlicePtr() {
	var target *[]time.Duration

	d, _ := time.ParseDuration("300ms")

	out, err := Convert(target, []string{"300ms"})
	s.Nil(err)
	s.Equal([]time.Duration{d}, *out.(*[]time.Duration))

	x := []string{"300ms"}
	out, err = Convert(target, &x)
	s.Nil(err)
	s.Equal([]time.Duration{d}, *out.(*[]time.Duration))
}

func (s *SuiteConvertTester) Test_ConvertWithReflect() {
	type User struct {
		Name string
	}

	user := User{Name: "foo"}

	var target User
	out, err := Convert(target, user)
	s.Nil(err)
	s.Equal(user, out.(User))

	var targetPtr *User
	_, err = Convert(targetPtr, user)
	s.NotNil(err)

	var targetInt int
	_, err = Convert(targetInt, "1.23abc")
	s.NotNil(err)
}

func TestSuiteConvert(t *testing.T) {
	suite.Run(t, new(SuiteConvertTester))
}
