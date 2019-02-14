package reflection

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZeroValueExportedStructFieldNames(t *testing.T) {
	type SubStruct struct {
		Int        int
		IntPtr     *int
		IntZero    int
		IntPtrZero *int
	}

	type Struct struct {
		ignore bool
		Ignore bool `tag:"-"`

		Int        int
		IntPtr     *int
		IntZero    int `tag:"intZero"`
		IntPtrZero *int

		Slice     []int `tag:"slice,omitempty"`
		SliceZero []int

		Sub SubStruct
	}

	st := Struct{
		Int:    666,
		IntPtr: new(int),
		Slice:  []int{1, 2, 0, 4, 5, 0},
		Sub: SubStruct{
			Int:    666,
			IntPtr: new(int),
		},
	}

	expected := []string{
		"prefix.Ignore",
		"prefix.IntZero",
		"prefix.IntPtrZero",
		"prefix.Slice[2]",
		"prefix.Slice[5]",
		"prefix.SliceZero",
		"prefix.Sub.IntZero",
		"prefix.Sub.IntPtrZero",
	}

	zeroNames := ZeroValueExportedStructFieldNames(st, "prefix.", "")
	assert.ElementsMatch(t, expected, zeroNames)

	expectedWithTag := []string{
		"prefix.intZero",
		"prefix.IntPtrZero",
		"prefix.slice[2]",
		"prefix.slice[5]",
		"prefix.SliceZero",
		"prefix.Sub.IntZero",
		"prefix.Sub.IntPtrZero",
	}

	zeroNames = ZeroValueExportedStructFieldNames(st, "prefix.", "tag")
	assert.ElementsMatch(t, expectedWithTag, zeroNames)

	expectedWithIgnore := []string{
		"prefix.IntZero",
		"prefix.Slice[2]",
		"prefix.Slice[5]",
		"prefix.Sub.IntZero",
	}

	zeroNames = ZeroValueExportedStructFieldNames(st, "prefix.", "", "prefix.IntZero", "prefix.Slice", "prefix.Sub", "prefix.Sub.IntZero")
	t.Log(zeroNames)
	assert.ElementsMatch(t, expectedWithIgnore, zeroNames)
}
