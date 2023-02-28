package template

import (
	"fmt"
	"reflect"
	"testing"
)

type Person struct {
	Id   int64
	Name string
	Age  int
}

func RawFiledValues(ptr interface{}) ([]interface{}, error) {
	var resp []interface{}

	v := reflect.ValueOf(ptr)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ToMap only accepts structs; got %T", v)
	}

	for i := 0; i < v.NumField(); i++ {
		fileName := v.Type().Field(i).Name

		if fileName == "Id" {
			continue
		}
		value := v.FieldByName(fileName)
		resp = append(resp, value)

	}
	fmt.Println(resp...)
	return resp, nil
}

func TestGet(t *testing.T) {
	p := Person{Name: "罗剑锋", Id: 2}

	RawFiledValues(&p)

}
