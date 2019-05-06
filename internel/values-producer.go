package internel

// ValuesProducer 多列提供者
// 用于提供每一行扫描时所需的多列的实体
// 如
// var vp ValuesProducer
// for rows.Next(){
//     rows.Scan(vp.Values()...)
// }
//
type ValuesProducer interface {
	Values() []interface{}
}

type valuesProducer struct {
	columns     []Column
	rowProducer RowProducer
	cache       []interface{}
}

func (s *valuesProducer) Values() []interface{} {
	row := s.rowProducer()
	for i, p := range s.columns {
		value, _ := p.PointerOf(row)
		s.cache[i] = value.Interface()
	}
	return s.cache
}
