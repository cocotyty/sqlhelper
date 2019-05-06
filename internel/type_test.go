package internel

import (
	"reflect"
	"testing"
)

type testTypeInfoFactoryStruct struct{}

var testTypeInfoFactoryStr string

var ptrTotestTypeInfoFactoryStr = &testTypeInfoFactoryStr

var testTypeInfoFactoryTable = []struct {
	value interface{}
	typ   SupportType
	elm   reflect.Type
}{
	{&testTypeInfoFactoryStr, TypeRawType, reflect.TypeOf("")},
	{&[]byte{}, TypeRawType, reflect.TypeOf([]byte{})},
	{&[]string{}, TypeSliceOfRawType, reflect.TypeOf("")},
	{&[][]byte{}, TypeSliceOfRawType, reflect.TypeOf([]byte{})},
	{&[]testTypeInfoFactoryStruct{}, TypeSliceOfStruct, reflect.TypeOf(testTypeInfoFactoryStruct{})},
	{&[]*testTypeInfoFactoryStruct{}, TypeSliceOfPtrToStruct, reflect.TypeOf(testTypeInfoFactoryStruct{})},
	{&[]*string{}, TypeSliceOfPtrToRawType, reflect.TypeOf("")},
	{&[]*[]byte{}, TypeSliceOfPtrToRawType, reflect.TypeOf([]byte{})},
}

var testTypeInfoFactoryErrorTable = []interface{}{
	&[]*[][]byte{},
	&ptrTotestTypeInfoFactoryStr,
}

func TestTypeInfoFactory_Get(t *testing.T) {
	factory := NewTypeInfoFactory()
	for index, test := range testTypeInfoFactoryTable {
		info, err := factory.Get(test.value)
		if err != nil {
			t.Fatal(index, test, err)
		}
		if info.Type != test.typ {
			t.Fatal(index, test, info)
		}
		if info.ElemType != test.elm {
			t.Fatal(index, test, info)
		}
	}

	for index, test := range testTypeInfoFactoryErrorTable {
		_, err := factory.Get(test)
		if err == nil {
			t.Fatal(index, "must be error")
		}
	}
}
