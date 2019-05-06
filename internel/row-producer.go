package internel

import "reflect"

type RowProducer func() reflect.Value

func sliceRP(v reflect.Value) RowProducer {
	// 创建一个非空的slice
	if v.Elem().Len() == 0 {
		v.Elem().Set(reflect.MakeSlice(v.Elem().Type(), 0, 15))
	}
	// 拿到每一行的类型
	rowType := v.Type().Elem().Elem()
	slice := v.Elem()
	return func() reflect.Value {
		if slice.Len() < slice.Cap() {
			slice.Set(slice.Slice(0, slice.Len()+1))
			return slice.Index(slice.Len() - 1).Addr()
		}
		created := reflect.New(rowType)
		v.Elem().Set(reflect.Append(v.Elem(), created.Elem()))
		return v.Elem().Index(v.Elem().Len() - 1).Addr()
	}
}

func sliceOfPtrRP(v reflect.Value) RowProducer {
	slice := v.Elem()
	if slice.Len() == 0 {
		slice.Set(reflect.MakeSlice(slice.Type(), 0, 15))
	}
	rowType := v.Type().Elem().Elem().Elem()
	return func() reflect.Value {
		created := reflect.New(rowType)
		slice.Set(reflect.Append(slice, created))
		return created
	}
}

func structRP(v reflect.Value) RowProducer {
	return func() reflect.Value {
		return v.Elem()
	}
}

func rawTypeRP(v reflect.Value) RowProducer {
	return func() reflect.Value {
		return v
	}
}
