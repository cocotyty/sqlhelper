package internel

import (
	"testing"
)

var snakeTestTable = []struct {
	Source      string
	Destination string
}{
	{"AbcDe", "abc_de"},
	{"aB", "a_b"},
	{"aBB", "a_bb"},
	{"iQ", "i_q"},
	{"i", "i"},
	{"K", "k"},
}

func TestToSnake(t *testing.T) {
	for _, testData := range snakeTestTable {
		result := ToSnake(testData.Source)
		if result != testData.Destination {
			t.Fatal("expect from " + testData.Source + " to " + testData.Destination + " but get " + result)
		}
	}
}
