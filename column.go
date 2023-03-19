package main

import (
	"fmt"
	"strings"
)

type Column struct {
	Name      string
	Type      string
	Nullable  string
	Key       string
	f5        interface{}
	increment string
}

func (dc *Column) structFieldName() string {
	return strings.Title(dc.Name)
}

func (dc *Column) structFieldType() string {

	t := strings.Replace(dc.Type, " unsigned", "", -1)

	if strings.HasPrefix(t, "varchar(") {
		return "string"
	}

	if strings.HasPrefix(t, "decimal(") {
		return "float32"
	}

	if strings.HasPrefix(t, "char(") {
		s := reNumberBetweenBrackets.FindString(t)
		return fmt.Sprintf("[%s]byte", s[1:len(s)-1])
	}

	switch t {
	case "text", "longtext", "mediumtext":
		return "string"
	case "bigint", "int", "tinyint", "mediumint":
		return "int"
	case "tinyint(1)":
		return "bool"
	case "timestamp":
		return "int"
	case "datetime", "date":
		return "sql.NullTime"
	case "json":
		return "[]byte"
	}

	return t
}
