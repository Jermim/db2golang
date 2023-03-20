package main

import (
	"database/sql"
	"errors"
	"strings"
)

type DBFieldDescription struct {
	Name      string
	Type      string
	Nullable  string
	Key       string
	f5        interface{}
	increment string
}

type Column struct {
	descr      *DBFieldDescription
	Name       string
	Type       string
	PrimaryKey bool
	Nullable   bool
}

func (dc *Column) Scan(rows *sql.Rows) error {
	if dc.descr != nil {
		return errors.New("Column::scan already called")
	}

	dc.descr = &DBFieldDescription{}
	fields := []interface{}{&dc.descr.Name, &dc.descr.Type, &dc.descr.Nullable, &dc.descr.Key, &dc.descr.f5, &dc.descr.increment}

	if err := rows.Scan(fields...); err != nil {
		return err
	}

	dc.init()

	return nil
}

func (dc *Column) init() {
	dc.Name = strings.Title(dc.descr.Name)
	dc.Type = structFieldType(dc)
	dc.PrimaryKey = dc.descr.Key == "PRI"
	dc.Nullable = dc.descr.Nullable == "YES"
}

func (dc *Column) DBName() string {
	return dc.descr.Name
}

func (dc *Column) DBType() string {
	return dc.descr.Type
}

func (dc *Column) stringType() string {
	if dc.Nullable {
		return "sql.NullString"
	}

	return "string"
}

func (dc *Column) intType() string {
	if dc.Nullable {
		return "sql.NullInt32"
	}

	return "int"
}

func (dc *Column) boolType() string {
	if dc.Nullable {
		return "sql.NullBool"
	}

	return "bool"
}

func (dc *Column) floatType() string {
	if dc.Nullable {
		return "sql.NullFloat64"
	}

	return "float64"
}

func structFieldType(dc *Column) string {

	t := strings.Replace(dc.descr.Type, " unsigned", "", -1)

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
