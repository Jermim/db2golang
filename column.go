package main

import (
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

func (dc *Column) isPrimaryKey() bool {
	return dc.Key == "PRI"
}

func (dc *Column) isNullable() bool {
	return dc.Nullable == "YES"
}

func (dc *Column) stringType() string {
	if dc.isNullable() {
		return "sql.NullString"
	}

	return "string"
}

func (dc *Column) intType() string {
	if dc.isNullable() {
		return "sql.NullInt32"
	}

	return "int"
}

func (dc *Column) boolType() string {
	if dc.isNullable() {
		return "sql.NullBool"
	}

	return "bool"
}

func (dc *Column) floatType() string {
	if dc.isNullable() {
		return "sql.NullFloat64"
	}

	return "float64"
}

func (dc *Column) structFieldType() string {

	t := strings.Replace(dc.Type, " unsigned", "", -1)

	if strings.HasPrefix(t, "varchar(") || strings.HasPrefix(t, "char(") {
		return dc.stringType()
	}

	if strings.HasPrefix(t, "decimal(") {
		return dc.floatType()
	}

	switch t {
	case "text", "longtext", "mediumtext":
		return dc.stringType()
	case "bigint", "int", "tinyint", "mediumint":
		return dc.intType()
	case "tinyint(1)":
		return dc.boolType()
	case "timestamp", "datetime":
		return "[]uint8"
	case "date":
		return "[]uint8"
	case "json":
		return "[]byte"
	}

	return t
}
