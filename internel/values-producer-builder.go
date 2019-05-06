package internel

import (
	"reflect"
)

func NewValuesProducerBuilder(fieldProducer *TypeFieldProducer) *ValuesProducerBuilder {
	return &ValuesProducerBuilder{
		fieldProducer:   fieldProducer,
		typeInfoFactory: NewTypeInfoFactory(),
	}
}

type ValuesProducerBuilder struct {
	fieldProducer   *TypeFieldProducer
	typeInfoFactory *TypeInfoFactory
}

func (builder *ValuesProducerBuilder) Build(obj interface{}, columnNames []string) (ValuesProducer, bool, error) {
	info, err := builder.typeInfoFactory.Get(obj)
	if err != nil {
		return nil, false, err
	}

	value := reflect.ValueOf(obj)

	rowProducer := info.Type.RowProducer(value)

	columns := builder.GetColumns(info, columnNames)

	return &valuesProducer{
		columns:     columns,
		rowProducer: rowProducer,
		cache:       make([]interface{}, len(columns)),
	}, info.Type == TypeRawType || info.Type == TypeStruct, nil
}

func (builder *ValuesProducerBuilder) GetColumns(info *TypeInfo, columnNames []string) (columns []Column) {
	columns = make([]Column, 0, len(columnNames))
	switch info.Type {
	case TypeStruct, TypeSliceOfPtrToStruct, TypeSliceOfStruct:
		// struct columns
		fields := builder.fieldProducer.Fields(info.ElemType)
		for _, col := range columnNames {
			field, ok := fields[col]
			if ok {
				columns = append(columns, field)
				continue
			}
			columns = append(columns, ignoreRowColumn)
		}
	case TypeRawType, TypeSliceOfPtrToRawType, TypeSliceOfRawType:
		// RawType columns
		for i := range columnNames {
			if i == 0 {
				columns = append(columns, rawTypeColumn)
				continue
			}
			columns = append(columns, ignoreRowColumn)
		}
	}
	return
}
