package internel

import (
	"reflect"
	"testing"
)

func TestSliceOfPtrRP(t *testing.T) {
	var a []*int
	typ := sliceOfPtrRP(reflect.ValueOf(&a))()
	if typ.Kind() != reflect.Ptr {
		t.Fatal(typ.Kind())
	}
	if typ.Elem().Kind() != reflect.Int {
		t.Fatal()
	}
	typ.Elem().SetInt(3)
	if *a[0] != 3 {
		t.Fatal()
	}

	var x []*testA
	producer := sliceOfPtrRP(reflect.ValueOf(&x))
	val := producer()
	val.Elem().FieldByName("A1").SetString("string")
	if x[0].A1 != "string" {
		t.Fatal("need string", x[0])
	}

	val = producer()
	val.Elem().FieldByName("A1").SetString("string2")
	if x[1].A1 != "string2" {
		t.Fatal("need string2", x[1])
	}
}

func TestSliceRP(t *testing.T) {
	var b []int
	val := sliceRP(reflect.ValueOf(&b))()
	if val.Kind() != reflect.Ptr {
		t.Fatal(val.Kind())
	}
	if val.Elem().Kind() != reflect.Int {
		t.Fatal(val.Kind())
	}
	val.Elem().SetInt(2)
	if b[0] != 2 {
		t.Fatal()
	}

	var x []testA
	producer := sliceRP(reflect.ValueOf(&x))
	val = producer()
	val.Elem().FieldByName("A1").SetString("string")
	if x[0].A1 != "string" {
		t.Fatal("need string", x[0])
	}

	val = producer()
	val.Elem().FieldByName("A1").SetString("string2")
	if x[1].A1 != "string2" {
		t.Fatal("need string2", x[1])
	}
}

func TestStructRP(t *testing.T) {
	type typ struct {
		A int
	}
	var v = typ{A: 999}

	value := structRP(reflect.ValueOf(&v))()

	if value.Interface().(typ).A != 999 {
		t.Fatal()
	}
	value.Field(0).SetInt(888)
	if v.A != 888 {
		t.Fatal()
	}
}

func TestRawTypeRP(t *testing.T) {
	v := 999

	value := rawTypeRP(reflect.ValueOf(&v))()

	if *(value.Interface().(*int)) != 999 {
		t.Fatal()
	}
	value.Elem().SetInt(777)
	if v != 777 {
		t.Fatal()
	}
}
