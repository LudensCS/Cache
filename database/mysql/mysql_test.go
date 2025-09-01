package mysql

import (
	"fmt"
	"testing"
)

func TestSQL(t *testing.T) {
	db, err := Register("ludens:123456@tcp(127.0.0.1:3306)/itcast?charset=utf8mb4&loc=Local&parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	var datas = []*Data{
		{Key: "Jack", Value: []byte("male")},
		{Key: "Lucy", Value: []byte("female")},
		{Key: "David", Value: []byte("male")},
	}
	if err := Insert(db, datas); err != nil {
		t.Fatal(err)
	}
	rows, err := Select(db, "*")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) == 0 {
		t.Fatal("empty returns")
	}
	for _, data := range rows {
		if data.Key != "Jack" && data.Key != "Lucy" && data.Key != "David" {
			t.Fatal(fmt.Errorf("Select Rows Error"))
		}
		if string(data.Value) != "male" && string(data.Value) != "female" {
			t.Fatal(fmt.Errorf("Select Rows Error"))
		}
	}
	row, err := Select(db, "Jack")
	if err != nil {
		t.Fatal(err)
	}
	if len(row) == 0 {
		t.Fatal("empty return")
	}
	if row[0].Key != "Jack" || string(row[0].Value) != "male" {
		t.Fatal("Select Row Error")
	}
}
