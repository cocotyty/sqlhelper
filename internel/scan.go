package internel

import "database/sql"

type SQLRows interface {
	Next() bool
	Columns() ([]string, error)
	Scan(dest ...interface{}) error
	Close() error
}

var GlobalTypeFieldProducer = NewTypeFieldProducer(SnakeMapper)

var GlobalScanner = &RowsScanner{
	builder: NewValuesProducerBuilder(GlobalTypeFieldProducer),
}

func Scan(rows SQLRows, ptr interface{}) (err error) {
	return GlobalScanner.Scan(rows, ptr)
}

type RowsScanner struct {
	builder *ValuesProducerBuilder
}

func (rs *RowsScanner) SetMapper(mapper Mapper) {
	rs.builder.fieldProducer.Mapper = mapper
}

func (rs *RowsScanner) Scan(rows SQLRows, ptr interface{}) (err error) {
	cols, err := rows.Columns()
	if err != nil {
		rows.Close()
		return
	}
	producer, oneLine, err := rs.builder.Build(ptr, cols)
	if err != nil {
		rows.Close()
		return
	}
	rowsSize := 0
	for rows.Next() {
		rowsSize++
		err = rows.Scan(producer.Values()...)
		if err != nil {
			rows.Close()
			return
		}
		if oneLine {
			break
		}
	}
	rows.Close()
	if oneLine && rowsSize == 0 {
		return sql.ErrNoRows
	}
	return nil
}
