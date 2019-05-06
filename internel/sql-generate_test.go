package internel

import (
	"reflect"
	"testing"
	"time"
)

func TestGenerateInsertSQL(t *testing.T) {
	type testStruct struct {
		ID   int
		Name string
	}
	type testStruct2 struct {
		ID          int
		Name        string
		ShowName    string
		CreatedTime time.Time
	}
	fields := toNamedFields(Fields(reflect.TypeOf(&testStruct{}), SnakeMapper))

	sql, _, _ := GenerateInsertSQL("table", fields, true)
	if sql != "INSERT INTO `table` (`name`) VALUES (?)" {
		t.Fatal(sql)
	}

	fields = toNamedFields(Fields(reflect.TypeOf(&testStruct2{}), SnakeMapper))
	sql, _, list := GenerateInsertSQL("table", fields, true)
	if sql != "INSERT INTO `table` (`name`,`show_name`,`created_time`) VALUES (?,?,?)" {
		t.Fatal(sql)
	}
	if list[0].i != 1 && list[1].i != 2 && list[2].i != 3 {
		t.Fatal(list[0].i, list[1].i, list[2].i)
	}

	fields = toNamedFields(Fields(reflect.TypeOf(&testStruct2{}), SnakeMapper))
	sql, _, list = GenerateInsertSQL("table", fields, false)
	if sql != "INSERT INTO `table` (`id`,`name`,`show_name`,`created_time`) VALUES (?,?,?,?)" {
		t.Fatal(sql)
	}
	if list[0].i != 0 && list[1].i != 1 && list[2].i != 2 {
		t.Fatal(list[0].i, list[1].i, list[2].i)
	}
}

func TestGenerateSelectSQL(t *testing.T) {

	type testStruct struct {
		ID          int
		Name        string
		ShowName    string
		CreatedTime time.Time
	}
	fields := toNamedFields(Fields(reflect.TypeOf(&testStruct{}), SnakeMapper))

	sql:= GenerateSelectSQL("table", fields)
	if sql != "SELECT `id`,`name`,`show_name`,`created_time` FROM `table` " {
		t.Fatal(sql)
	}
}

func TestGenerateUpdateSQL(t *testing.T) {
	type testStruct struct {
		ID   int
		Name string
	}
	type testStruct2 struct {
		ID          int
		Name        string
		ShowName    string
		CreatedTime time.Time
	}
	fields := toNamedFields(Fields(reflect.TypeOf(&testStruct{}), SnakeMapper))

	sql, _ := GenerateUpdateSQL("table", fields)
	if sql != "UPDATE `table` SET `name`=?" {
		t.Fatal(sql)
	}

	fields = toNamedFields(Fields(reflect.TypeOf(&testStruct2{}), SnakeMapper))
	sql, list := GenerateUpdateSQL("table", fields)
	if sql != "UPDATE `table` SET `name`=?,`show_name`=?,`created_time`=?" {
		t.Fatal(sql)
	}
	if list[0].i != 1 && list[1].i != 2 && list[2].i != 3 {
		t.Fatal(list[0].i, list[1].i, list[2].i)
	}
}

func TestSQLGenerator_MapTable(t *testing.T) {
	type testStruct2 struct {
		ID          int
		Name        string
		ShowName    string
		CreatedTime time.Time
	}
	sg := NewSQLGenerator(NewTypeFieldProducer(SnakeMapper))
	sg.MapTable("not_type_name", testStruct2{})
	ts := &testStruct2{
		Name:        "thisisname",
		CreatedTime: time.Now(),
	}
	sql, args, _ := sg.PrepareInsert(ts)
	if sql != "INSERT INTO `not_type_name` (`name`,`show_name`,`created_time`) VALUES (?,?,?)" {
		t.Fatal(sql)
	}
	if args[0] != "thisisname" && args[1] != "" && args[2] != ts.CreatedTime {
		t.Fatal(args)
	}
}

func TestSQLGenerator_PrepareInsert(t *testing.T) {
	type ThisUseTypeName struct {
		Id          int
		Name        string
		ShowName    string
		CreatedTime time.Time
	}
	ts := &ThisUseTypeName{
		Name:        "thisisname",
		CreatedTime: time.Now(),
	}
	sg := NewSQLGenerator(NewTypeFieldProducer(SnakeMapper))
	sql, args, _ := sg.PrepareInsert(ts)
	if sql != "INSERT INTO `this_use_type_name` (`name`,`show_name`,`created_time`) VALUES (?,?,?)" {
		t.Fatal(sql)
	}
	if args[0] != "thisisname" && args[1] != "" && args[2] != ts.CreatedTime {
		t.Fatal(args)
	}

	ts.Id = 5
	sql, args, _ = sg.PrepareInsert(ts)
	if sql != "INSERT INTO `this_use_type_name` (`id`,`name`,`show_name`,`created_time`) VALUES (?,?,?,?)" {
		t.Fatal(sql)
	}
	if args[0] != 5 && args[1] != "thisisname" && args[2] == "" && args[3] != ts.CreatedTime {
		t.Fatal(args)
	}
}

func TestSQLGenerator_PrepareUpdate(t *testing.T) {
	type ThisUseTypeName struct {
		Id          int
		Name        string
		ShowName    string
		CreatedTime time.Time
	}
	ts := &ThisUseTypeName{
		Name:        "thisisname",
		CreatedTime: time.Now(),
	}
	sg := NewSQLGenerator(NewTypeFieldProducer(SnakeMapper))
	sql, args, _ := sg.PrepareUpdate(ts)
	if sql != "UPDATE `this_use_type_name` SET `name`=?,`show_name`=?,`created_time`=?" {
		t.Fatal(sql)
	}
	if args[0] != "thisisname" && args[1] != "" && args[2] != ts.CreatedTime {
		t.Fatal(args)
	}
}

func TestSQLGenerator_PrepareUpdateByID(t *testing.T) {
	type ThisUseTypeName struct {
		Id          int
		Name        string
		ShowName    string
		CreatedTime time.Time
	}
	ts := &ThisUseTypeName{
		Name:        "thisisname",
		CreatedTime: time.Now(),
	}
	sg := NewSQLGenerator(NewTypeFieldProducer(SnakeMapper))
	sql, args, _ := sg.PrepareUpdate(ts)
	if sql != "UPDATE `this_use_type_name` SET `name`=?,`show_name`=?,`created_time`=?" {
		t.Fatal(sql)
	}
	if args[0] != "thisisname" && args[1] != "" && args[2] != ts.CreatedTime {
		t.Fatal(args)
	}
}

func BenchmarkSQLGenerator_PrepareInsert(b *testing.B) {
	type ThisUseTypeName struct {
		Id          int
		Name        string
		ShowName    string
		CreatedTime time.Time
	}
	ts := &ThisUseTypeName{
		Name:        "thisisname",
		CreatedTime: time.Now(),
	}
	sg := NewSQLGenerator(NewTypeFieldProducer(SnakeMapper))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sg.PrepareInsert(ts)
		}
	})
}
