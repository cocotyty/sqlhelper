package internel

import (
	"database/sql"
	"reflect"
	"sync"
	"time"
)

type SupportType int

const (
	TypeStruct  SupportType = iota
	TypeRawType             // rawType 包括 bool int float string []byte time.Time 和所有scanner
	TypeSliceOfStruct
	TypeSliceOfRawType
	TypeSliceOfPtrToStruct
	TypeSliceOfPtrToRawType
)

func (typ SupportType) String() string {
	switch typ {
	case TypeStruct:
		return "TypeStruct"
	case TypeRawType:
		return "TypeRawType"
	case TypeSliceOfStruct:
		return "TypeSliceOfStruct"
	case TypeSliceOfRawType:
		return "TypeSliceOfRawType"
	case TypeSliceOfPtrToStruct:
		return "TypeSliceOfPtrToStruct"
	case TypeSliceOfPtrToRawType:
		return "TypeSliceOfPtrToRawType"
	default:
		return ""
	}
}

func (typ SupportType) RowProducer(value reflect.Value) (p RowProducer) {
	switch typ {
	case TypeStruct:
		p = structRP(value)
	case TypeRawType:
		p = rawTypeRP(value)
	case TypeSliceOfPtrToStruct, TypeSliceOfPtrToRawType:
		p = sliceOfPtrRP(value)
	case TypeSliceOfStruct, TypeSliceOfRawType:
		p = sliceRP(value)
	}
	return p
}

type TypeInfo struct {
	Type     SupportType
	ElemType reflect.Type
}

var (
	bytesType = reflect.TypeOf([]byte{})
	timeType  = reflect.TypeOf(time.Time{})
)

func NewTypeInfoFactory() *TypeInfoFactory {
	return &TypeInfoFactory{
		cache: make(map[reflect.Type]*TypeInfo),
	}
}

type TypeInfoFactory struct {
	cache map[reflect.Type]*TypeInfo
	mutex sync.RWMutex
}

func (f *TypeInfoFactory) Get(obj interface{}) (info *TypeInfo, err error) {
	typ := reflect.TypeOf(obj)
	if typ.Kind() != reflect.Ptr {
		return nil, ErrInvalidScanType
	}

	f.mutex.RLock()
	if info, ok := f.cache[typ]; ok {
		f.mutex.RUnlock()
		return info, nil
	}
	f.mutex.RUnlock()

	info, _ = getInfo(typ)
	if info == nil {
		return nil, ErrInvalidScanType
	}
	f.mutex.Lock()
	f.cache[typ] = info
	f.mutex.Unlock()
	return
}

func getInfo(typ reflect.Type) (info *TypeInfo, pointer bool) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		pointer = true
	}

	info = &TypeInfo{}
	// 过滤掉time 和 bytes
	switch typ {
	case bytesType, timeType:
		info.Type = TypeRawType
		info.ElemType = typ
		return
	}
	// 类型断言是否为scanner
	if _, ok := typ.(sql.Scanner); ok {
		info.Type = TypeRawType
		info.ElemType = typ
		return
	}
	// 根据类别判断:
	switch typ.Kind() {
	case reflect.Slice:
		subInfo, isPointer := getInfo(typ.Elem())
		if subInfo == nil {
			return nil, false
		}
		switch subInfo.Type {
		case TypeRawType:
			if isPointer {
				info.Type = TypeSliceOfPtrToRawType
			} else {
				info.Type = TypeSliceOfRawType
			}
		case TypeStruct:
			if isPointer {
				info.Type = TypeSliceOfPtrToStruct
			} else {
				info.Type = TypeSliceOfStruct
			}
		default:
			return nil, false
		}
		info.ElemType = subInfo.ElemType

	case reflect.Struct:
		info.Type = TypeStruct
		info.ElemType = typ
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
		info.Type = TypeRawType
		info.ElemType = typ
	default:
		return nil, false
	}
	return
}
