package internel

import (
	"database/sql"
	_ "database/sql"
	"reflect"
	"unsafe"
)

//go:linkname convertAssign database/sql.convertAssign
func convertAssign(dest, src interface{}) error

// 代表列信息
type Column interface {
	// 获取当前行的该列的指针
	PointerOf(v reflect.Value) (val reflect.Value, err error)
}

var ignoreRowColumn = IgnoredColumn{}

type IgnoredColumn struct{}

// PointerOf 返回可以扫描的空值 不管传入任何
func (IgnoredColumn) PointerOf(v reflect.Value) (val reflect.Value, err error) {
	var a interface{}
	return reflect.ValueOf(&a), nil
}

var rawTypeColumn = RawTypeColumn{}

type RawTypeColumn struct{}

// PointerOf 返回值本身 这里需要保证由上游调用保证v是指向rawType值的指针 不做检查
func (RawTypeColumn) PointerOf(v reflect.Value) (val reflect.Value, err error) {
	return v, nil
}

// 表示结构体映射的查询返回字段
type Field struct {
	i     int          // 索引
	typ   reflect.Type //类型
	embed *Field
}

var sqlScannerType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

func (p *Field) PointerOf(v reflect.Value) (val reflect.Value, err error) {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		err = ErrInvalidScanType
		return
	}
	f := v.Field(p.i)
	if p.embed != nil {
		if p.typ.Kind() == reflect.Ptr {
			if f.IsNil() {
				value := reflect.NewAt(p.typ.Elem(), unsafe.Pointer(f.UnsafeAddr()))
				return p.embed.PointerOf(value.Elem())
			} else {
				return p.embed.PointerOf(f.Elem())
			}
		}
		return p.embed.PointerOf(f)
	}

	if p.typ.Kind() == reflect.Ptr {
		if p.typ.Implements(sqlScannerType) {
			return f.Addr(), nil
		}
		return reflect.ValueOf(&NullValue{
			Value: f,
		}), nil
	}
	return f.Addr(), nil
}

type NullValue struct {
	Value reflect.Value
}

func (n *NullValue) Scan(value interface{}) error {
	if value == nil {
		if !n.Value.IsNil() {
			n.Value.Set(reflect.Zero(n.Value.Type()))
		}
		return nil
	}
	val := n.Value
	typ := val.Type()
	for typ.Kind() == reflect.Ptr {
		subType := typ.Elem()
		if val.IsNil() {
			val.Set(reflect.New(subType))
		}
		typ = subType
		val = val.Elem()
	}
	err := convertAssign(val.Addr().Interface(), value)
	return err
}

// typ elem's type must be struct
func Fields(typ reflect.Type, mapper Mapper) map[string]*Field {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	fields := map[string]*Field{}
	num := typ.NumField()
	for i := 0; i < num; i++ {
		f := typ.Field(i)
		if f.Anonymous {
			next := Fields(f.Type, mapper)
			for name, path := range next {
				fields[name] = &Field{
					i:     i,
					typ:   f.Type,
					embed: path,
				}
			}
		}
		// 魔术操作 用于判断是否为私有属性 unexported
		if f.PkgPath != "" {
			continue
		}
		fieldMapperName := mapper(f.Name, f.Tag)
		fields[fieldMapperName] = &Field{i: i, typ: f.Type}
	}
	return fields
}
