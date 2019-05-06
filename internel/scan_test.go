package internel

import (
	"context"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
)

type testRowsStruct struct {
	A string `db:"a"`
	B string `db:"b"`
	C int    `db:"c"`
	D int    `db:"d"`
}

func TestRowsScanner_Scan(t *testing.T) {
	db, mock, _ := sqlmock.New()
	// 测试单个对象
	// 此处测试了类型转换
	mock.ExpectQuery("query1").WillReturnRows(
		sqlmock.NewRows([]string{"a", "b", "c"}).
			AddRow("1", 1, 1).
			AddRow("2", 2, 2),
	).WillReturnError(nil)
	rows, err := db.QueryContext(context.Background(), "query1")
	if err != nil {
		panic(err)
	}
	r := &testRowsStruct{
		D: 9,
	}
	err = GlobalScanner.Scan(rows, r)
	if err != nil {
		t.Fatal(err)
	}
	if r.A != "1" || r.B != "1" || r.C != 1 {
		t.Fatal(r)
	}
	if r.D != 9 {
		t.Fatal(r)
	}
	t.Log(r)

	// 测试指针类型list返回
	mock.ExpectQuery("query1").WillReturnRows(
		sqlmock.NewRows([]string{"a", "b", "c"}).
			AddRow("1", 1, 1).
			AddRow("2", 2, 2).
			AddRow("3", 3, 3).
			AddRow("4", 4, 4).
			AddRow("5", 5, 5).
			AddRow("6", 6, 6),
	).WillReturnError(nil)
	rows, err = db.QueryContext(context.Background(), "query1")
	if err != nil {
		panic(err)
	}
	var list []*testRowsStruct
	err = GlobalScanner.Scan(rows, &list)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 6 {
		t.Fatal(list)
	}
	t.Log(list[0], list[1], list[2], list[3], list[4], list[5])

	// 测试非指针类型list返回
	mock.ExpectQuery("query1").WillReturnRows(
		sqlmock.NewRows([]string{"a", "b", "c"}).
			AddRow("1", 1, 1).
			AddRow("2", 2, 2).
			AddRow("3", 3, 3).
			AddRow("4", 4, 4).
			AddRow("5", 5, 5).
			AddRow("6", 6, 6),
	).WillReturnError(nil)
	rows, err = db.QueryContext(context.Background(), "query1")
	if err != nil {
		panic(err)
	}
	var slice []testRowsStruct
	err = GlobalScanner.Scan(rows, &slice)
	if err != nil {
		t.Fatal(err)
	}
	if len(slice) != 6 {
		t.Fatal(slice)
	}
	t.Log(slice[0], slice[1], slice[2], slice[3], slice[4], slice[5])

	// 测试单个对象在columns信息多余情况下的问题
	// 此处测试了类型转换
	mock.ExpectQuery("query1").WillReturnRows(
		sqlmock.NewRows([]string{"a", "b", "c", "e"}).
			AddRow("1", 1, 1, 1),
	).WillReturnError(nil)
	rows, err = db.QueryContext(context.Background(), "query1")
	if err != nil {
		t.Fatal(err)
	}
	r = &testRowsStruct{
		D: 9,
	}
	err = GlobalScanner.Scan(rows, r)
	if err != nil {
		t.Fatal(err)
	}
	if r.A != "1" || r.B != "1" || r.C != 1 {
		t.Fatal(r)
	}
	if r.D != 9 {
		t.Fatal(r)
	}
	t.Log(r)

	// 测试单值
	mock.ExpectQuery("query1").WillReturnRows(
		sqlmock.NewRows([]string{"a", "b", "c", "e"}).
			AddRow("1", 1, 1, 1),
	).WillReturnError(nil)
	rows, err = db.QueryContext(context.Background(), "query1")
	if err != nil {
		t.Fatal(err)
	}
	value := 9
	err = GlobalScanner.Scan(rows, &value)
	if err != nil {
		t.Fatal(err)
	}
	if value != 1 {
		t.Fatal("expected 1 get ", value)
	}
	t.Log(value)

	// 测试bytes
	mock.ExpectQuery("query1").WillReturnRows(
		sqlmock.NewRows([]string{"a", "b", "c", "e"}).
			AddRow("1", 1, 1, 1),
	).WillReturnError(nil)
	rows, err = db.QueryContext(context.Background(), "query1")
	if err != nil {
		t.Fatal(err)
	}
	var bytesValue []byte
	err = GlobalScanner.Scan(rows, &bytesValue)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytesValue) != "1" {
		t.Fatal("expected 1 get ", string(bytesValue))
	}
	t.Log(bytesValue)
	// 测试单值多行的情况
	// 测试bytes
	mock.ExpectQuery("query1").WillReturnRows(
		sqlmock.NewRows([]string{"a", "b", "c", "e"}).
			AddRow("1", 1, 1, 1).
			AddRow("2", 2, 2, 2),
	).WillReturnError(nil)
	rows, err = db.QueryContext(context.Background(), "query1")
	if err != nil {
		t.Fatal(err)
	}
	var bytesList [][]byte
	err = GlobalScanner.Scan(rows, &bytesList)
	if err != nil {
		t.Fatal(err)
	}
	if string(bytesList[0]) != "1" {
		t.Fatal("expected 1 get ", string(bytesValue))
	}
	if string(bytesList[1]) != "2" {
		t.Fatal("expected 1 get ", string(bytesValue))
	}
	t.Log(bytesList)
}

type testReadOnlyRows struct {
	i    int
	cols []string
}

func (r *testReadOnlyRows) Next() bool {
	r.i++
	return r.i < 10
}

func (r *testReadOnlyRows) Columns() ([]string, error) {
	return r.cols, nil
}

func (r *testReadOnlyRows) Scan(dest ...interface{}) error {
	return nil
}

func (*testReadOnlyRows) Close() error {
	return nil
}

/**
go test -bench ^BenchmarkRowsScanner_Scan$
	goos: linux
	goarch: amd64
	pkg: internel
	BenchmarkRowsScanner_Scan-8   	  300000	      4833 ns/op
	PASS
	ok  	internel	1.504s
*/
func BenchmarkRowsScanner_Scan(b *testing.B) {
	cols := []string{"a", "b", "c", "d"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var list []testRowsStruct
		GlobalScanner.Scan(&testReadOnlyRows{cols: cols}, &list)
	}
}

/**
go test -bench ^BenchmarkRowsScanner_Scan2$
	goos: linux
	goarch: amd64
	pkg: internel
	BenchmarkRowsScanner_Scan2-8   	 1000000	      1359 ns/op
	PASS
	ok  	internel	1.379s
*/
func BenchmarkRowsScanner_Scan2(b *testing.B) {
	cols := []string{"a", "b", "c", "d"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var list []testRowsStruct
			GlobalScanner.Scan(&testReadOnlyRows{cols: cols}, &list)
		}
	})
}

func BenchmarkRowsScanner_Scan3(b *testing.B) {
	cols := []string{"a", "b", "c", "d"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var list []*testRowsStruct
			GlobalScanner.Scan(&testReadOnlyRows{cols: cols}, &list)
		}
	})
}
