package internel

import (
	"bytes"
	"reflect"
	"sort"
	"strings"
	"sync"
)

const idColumnName = "id"

var GlobalSQLGenerator = NewSQLGenerator(GlobalTypeFieldProducer)

type SQLGenerator struct {
	fieldProducer *TypeFieldProducer
	locker        sync.RWMutex
	tables        map[reflect.Type]tableInfo
}

func NewSQLGenerator(fieldProducer *TypeFieldProducer) *SQLGenerator {
	sqlGen := &SQLGenerator{
		fieldProducer: fieldProducer,
		tables:        map[reflect.Type]tableInfo{},
	}
	return sqlGen
}

// tableInfo 表预生成的信息
type tableInfo struct {
	Name         string
	IDField      *NamedField
	Insert       SqlPair
	InsertWithID SqlPair
	Update       SqlPair
	UpdateByID   SqlPair
	Select       string
}

type SqlPair struct {
	sql       string
	fieldArgs []*NamedField
}

func (s *SQLGenerator) mapTable(name string, typ reflect.Type) {
	ti := tableInfo{
		Name: name,
	}
	fields := toNamedFields(s.fieldProducer.Fields(typ))
	ti.Insert.sql, ti.IDField, ti.Insert.fieldArgs = GenerateInsertSQL(name, fields, true)
	ti.InsertWithID.sql, _, ti.InsertWithID.fieldArgs = GenerateInsertSQL(name, fields, false)
	ti.Update.sql, ti.Update.fieldArgs = GenerateUpdateSQL(name, fields)
	ti.UpdateByID.sql = ti.Update.sql + " WHERE `" + ti.IDField.Name + "` = ?"
	ti.UpdateByID.fieldArgs = append(ti.Update.fieldArgs, ti.IDField)
	ti.Select = GenerateSelectSQL(name, fields)
	s.locker.Lock()
	s.tables[typ] = ti
	s.locker.Unlock()
	return
}

// MapTable 关联表与类型
func (s *SQLGenerator) MapTable(name string, o interface{}) error {
	typ := reflect.TypeOf(o)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return ErrInvalidScanType
	}
	s.mapTable(name, typ)
	return nil
}

func toNamedFields(fields map[string]*Field) (list []*NamedField) {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return len(names[i]) < len(names[j])
	})

	list = make([]*NamedField, 0, len(names))

	for _, name := range names {
		list = append(list, &NamedField{Name: name, Field: *fields[name]})
	}
	return
}

// getStructValue 尝试获取结构体类型的值
func (s *SQLGenerator) getStructValue(o interface{}) (val reflect.Value, err error) {
	val = reflect.ValueOf(o)
	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		err = ErrInvalidScanType
		return
	}
	return
}

// getTableInfo 尝试获取表预生成的信息
func (s *SQLGenerator) getTableInfo(typ reflect.Type) (table tableInfo, err error) {
	s.locker.RLock()
	table, ok := s.tables[typ]
	s.locker.RUnlock()
	if !ok {
		name := TypeName(typ)
		s.mapTable(s.fieldProducer.Mapper(name, ""), typ)
		s.locker.RLock()
		table, _ = s.tables[typ]
		s.locker.RUnlock()
	}
	return
}

// 将字段信息与传入值 转换为最终查询时用的参数
func fieldsToArgs(value reflect.Value, fields []*NamedField) (args []interface{}, err error) {
	// 4 假定通常情况下查询修改等场景多余4个以内的参数
	args = make([]interface{}, 0, len(fields)+4)
	for _, arg := range fields {
		var argValue reflect.Value
		argValue, err = arg.PointerOf(value)
		if err != nil {
			return
		}
		args = append(args, argValue.Elem().Interface())
	}
	return
}

// PrepareUpdateByID 通过传入的o为UpdateByID操作准备SQL语句与参数。
func (s *SQLGenerator) PrepareUpdateByID(o interface{}) (sql string, args []interface{}, err error) {
	val, err := s.getStructValue(o)
	if err != nil {
		return
	}
	table, err := s.getTableInfo(val.Type())
	if err != nil {
		return
	}
	args, err = fieldsToArgs(val, table.UpdateByID.fieldArgs)
	if err != nil {
		return
	}
	return table.UpdateByID.sql, args, nil
}

// PrepareUpdate 通过传入的o为Update操作准备SQL语句与参数。
func (s *SQLGenerator) PrepareUpdate(o interface{}) (sql string, args []interface{}, err error) {
	val, err := s.getStructValue(o)
	if err != nil {
		return
	}
	table, err := s.getTableInfo(val.Type())
	if err != nil {
		return
	}
	args, err = fieldsToArgs(val, table.Update.fieldArgs)
	if err != nil {
		return
	}
	return table.Update.sql, args, nil
}

func (s *SQLGenerator) PrepareSelectFrom(o interface{}) (sql string, err error) {
	typ := reflect.TypeOf(o)
	for typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		err = ErrInvalidScanType
		return
	}

	info, err := s.getTableInfo(typ)
	if err != nil {
		return
	}
	sql = info.Select
	return
}

// PrepareInsert 通过传入的o为Insert操作准备SQL语句与参数。
func (s *SQLGenerator) PrepareInsert(o interface{}) (sql string, args []interface{}, err error) {
	val, err := s.getStructValue(o)
	if err != nil {
		return
	}
	table, err := s.getTableInfo(val.Type())
	if err != nil {
		return
	}
	pair := table.Insert
	// 检查是否需要插入ID
	idVal, err := table.IDField.PointerOf(val)
	idVal = idVal.Elem()
	switch idVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if idVal.Int() != 0 {
			pair = table.InsertWithID
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if idVal.Uint() != 0 {
			pair = table.InsertWithID
		}
	}
	args, err = fieldsToArgs(val, pair.fieldArgs)
	return pair.sql, args, nil
}

type NamedField struct {
	Name string
	Field
}

func GenerateInsertSQL(table string, fields []*NamedField, skipID bool) (sql string, id *NamedField, list []*NamedField) {
	buf := bytes.NewBuffer([]byte("INSERT INTO `"))
	buf.WriteString(table)
	buf.WriteString("` (")
	i := 0
	for _, field := range fields {
		if skipID {
			if strings.ToLower(field.Name) == idColumnName {
				id = field
				continue
			}
		}
		list = append(list, field)
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('`')
		buf.WriteString(field.Name)
		buf.WriteByte('`')
		i++
	}
	buf.WriteString(") VALUES (")

	for ; i > 0; i-- {
		buf.WriteByte('?')
		if i != 1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteString(`)`)
	return buf.String(), id, list
}

func GenerateUpdateSQL(table string, fields []*NamedField) (sql string, list []*NamedField) {
	buf := bytes.NewBuffer([]byte("UPDATE `"))
	buf.WriteString(table)
	buf.WriteString("` SET ")
	i := 0
	for _, field := range fields {
		if strings.ToLower(field.Name) == idColumnName {
			continue
		}
		list = append(list, field)

		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('`')
		buf.WriteString(field.Name)
		buf.WriteString("`=?")
		i++
	}
	return buf.String(), list
}

func GenerateSelectSQL(table string, fields []*NamedField) (sql string) {
	buf := bytes.NewBuffer([]byte("SELECT "))
	i := 0
	for _, field := range fields {
		if i != 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('`')
		buf.WriteString(field.Name)
		buf.WriteByte('`')
		i++
	}
	buf.WriteString(" FROM `")
	buf.WriteString(table)
	buf.WriteString("` ")
	return buf.String()
}

// 获取类型名称 不包含包名
func TypeName(typ reflect.Type) string {
	typeName := typ.Name()
	if typeName == "" {
		return ""
	}
	pos := strings.LastIndex(typeName, ".")
	if pos != -1 {
		typeName = typeName[pos:]
	}
	return typeName
}
