package internel

import (
	"reflect"
)

type Mapper func(name string, tag reflect.StructTag) string

func TagMapper(name string, tag reflect.StructTag) string {
	dbName, _ := tag.Lookup("db")
	return dbName
}

const x = 'A' - 'a'

func SnakeMapper(name string, tag reflect.StructTag) string {
	if dbName := TagMapper(name, tag); dbName != "" {
		return dbName
	}
	return ToSnake(name)
}

func ToSnake(str string) string {
	nameRunes := ([]rune)(str)
	buffer := make([]rune, 0, len(nameRunes)+2)
	lastIsUp := false
	for i, r := range nameRunes {
		if r >= 'A' && r <= 'Z' {
			if !lastIsUp && i != 0 {
				buffer = append(buffer, '_')
			}
			lastIsUp = true
			buffer = append(buffer, r-x)
		} else {
			lastIsUp = false
			buffer = append(buffer, r)
		}
	}
	return string(buffer)
}
