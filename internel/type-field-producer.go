package internel

import (
	"reflect"
	"sync"
)

type TypeFieldProducer struct {
	Mapper Mapper
	cache  map[reflect.Type]map[string]*Field
	locker sync.RWMutex
}

func NewTypeFieldProducer(mapper Mapper) *TypeFieldProducer {
	return &TypeFieldProducer{
		Mapper: mapper,
		cache:  map[reflect.Type]map[string]*Field{},
	}
}

func (p *TypeFieldProducer) SetMapper(mapper Mapper) {
	p.Mapper = mapper
}

func (p *TypeFieldProducer) Fields(typ reflect.Type) map[string]*Field {
	p.locker.RLock()
	fields, ok := p.cache[typ]
	if ok {
		p.locker.RUnlock()
		return fields
	}
	p.locker.RUnlock()

	fields = Fields(typ, p.Mapper)

	p.locker.Lock()
	p.cache[typ] = fields
	p.locker.Unlock()
	return fields
}
