package internel

import (
	"reflect"
	"testing"
	"time"
)

type testA struct {
	A1 string
	A2 int64
	A3 int
	a4 int64
	A5 *time.Time
	testB
}

type testC struct {
	C1 bool
}

type testB struct {
	B1 int
	B2 string
	*testC
}

// 测试Fields
func TestFields(t *testing.T) {
	a := reflect.TypeOf(&testA{})
	f := Fields(a, func(name string, tag reflect.StructTag) string {
		return name
	})

	if f["a4"] != nil {
		t.Fatal("未暴露属性不得存入字典")
	}
	// 测试正常属性
	val, err := f["A1"].PointerOf(reflect.ValueOf(&testA{A1: "teststring"}))
	if err != nil {
		t.Fatal(err)
	}
	if val.Elem().Interface() != "teststring" {
		t.Fatal("A1 excepectd ", "teststring", " get ", val.Elem().Interface())
	}

	// 测试普通嵌入属性
	val, err = f["B1"].PointerOf(reflect.ValueOf(&testA{testB: testB{
		B1: 1,
	}}))
	if err != nil {
		t.Fatal(err)
	}
	if val.Elem().Interface() != 1 {
		t.Fatal("B1 excepectd ", 1, " get ", val.Elem().Interface())
	}

	// 测试深层次指针嵌入属性
	val, err = f["C1"].PointerOf(reflect.ValueOf(&testA{testB: testB{
		testC: &testC{
			C1: true,
		},
	}}))
	if err != nil {
		t.Fatal(err)
	}
	if val.Elem().Interface() != true {
		t.Fatal("C1 excepectd ", true, " get ", val.Elem().Interface())
	}

	// 测试指针嵌入属性2 测试空指针是否出现panic情况
	val, err = f["C1"].PointerOf(reflect.ValueOf(&testA{testB: testB{}}))
	if err != nil {
		t.Fatal(err)
	}
	if val.Elem().Interface() != false {
		t.Fatal("C1 excepectd ", false, " get ", val.Elem().Interface())
	}
	// 测试指针
	val, err = f["A5"].PointerOf(reflect.ValueOf(&testA{}))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := val.Elem().Interface().(NullValue); !ok {
		t.Fatal(val.Elem().Interface())
	}
	before := time.Now()
	now := before

	val, err = f["A5"].PointerOf(reflect.ValueOf(&testA{A5: &now}))
	if err != nil {
		t.Fatal(err)
	}
	if nowValue, ok := val.Elem().Interface().(NullValue); !ok {
		t.Fatal(val.Elem().Interface())
	} else {
		nowValue.Scan(time.Now().Add(time.Hour + time.Minute))
		if now.Sub(before) < time.Hour {
			t.Fatal(now)
		}
	}

}

/*
    测试机器
		型号名称：	MacBook Pro
		型号标识符：	MacBookPro11,5
		处理器名称：	Intel Core i7
		处理器速度：	2.5 GHz
		处理器数目：	1
		核总数：	4
		L2 缓存（每个核）：	256 KB
		L3 缓存：	6 MB
		内存：	16 GB

	测试结果
		goos: darwin
		goarch: amd64
		pkg: git.pandatv.com/panda-public/mysql-go/internel
		10000000	       162 ns/op
		PASS
*/

// 测试取属性的速度
func BenchmarkField_PointerOf(b *testing.B) {
	f := Fields(reflect.TypeOf(&testA{}), func(name string, tag reflect.StructTag) string {
		return name
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f["C1"].PointerOf(reflect.ValueOf(&testA{testB: testB{}}))
	}
}

/*
    测试机器
		型号名称：	MacBook Pro
		型号标识符：	MacBookPro11,5
		处理器名称：	Intel Core i7
		处理器速度：	2.5 GHz
		处理器数目：	1
		核总数：	4
		L2 缓存（每个核）：	256 KB
		L3 缓存：	6 MB
		内存：	16 GB
	测试结果
		goos: darwin
		goarch: amd64
		pkg: git.pandatv.com/panda-public/mysql-go/internel
		1000000	      1643 ns/op
		PASS
*/
// 测试 初始化类型信息的速度
func BenchmarkFields(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Fields(reflect.TypeOf(&testA{}), func(name string, tag reflect.StructTag) string {
			return name
		})
	}
}

func TestIgnoredColumn_PointerOf(t *testing.T) {
	val, err := IgnoredColumn{}.PointerOf(reflect.ValueOf("x"))
	if err != nil {
		t.Fatal(err)
	}
	if !val.Elem().IsNil() {
		t.Fatal()
	}
}

func TestRawTypeColumn_PointerOf(t *testing.T) {
	val, err := RawTypeColumn{}.PointerOf(reflect.ValueOf("x"))
	if err != nil {
		t.Fatal(err)
	}
	if val.Interface().(string) != "x" {
		t.Fatal()
	}
}
