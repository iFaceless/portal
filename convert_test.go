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

func TestSuiteConvert(t *testing.T) {
	suite.Run(t, new(SuiteConvertTester))
}
