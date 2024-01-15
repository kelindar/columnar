package main

import (
	_ "embed"
	"os"
	"text/template"
)

//go:embed numbers.tpl
var numbers string

type Type struct {
	Name string
	Type string
}

func main() {
	t, err := template.New("numbers").Parse(numbers)
	if err != nil {
		panic(err)
	}

	dst, err := os.OpenFile("column_numbers.go", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	defer dst.Close()
	if err != nil {
		panic(err)
	}

	if err := t.Execute(dst, []Type{
		{Name: "Int", Type: "int"},
		{Name: "Int16", Type: "int16"},
		{Name: "Int32", Type: "int32"},
		{Name: "Int64", Type: "int64"},
		{Name: "Uint", Type: "uint"},
		{Name: "Uint16", Type: "uint16"},
		{Name: "Uint32", Type: "uint32"},
		{Name: "Uint64", Type: "uint64"},
		{Name: "Float32", Type: "float32"},
		{Name: "Float64", Type: "float64"},
	}); err != nil {
		panic(err)
	}
}
