package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	reNumberBetweenBrackets = regexp.MustCompile("(?si)(\\(\\d+\\))")
	reUnderscore            = regexp.MustCompile("(?is)_[a-z]")
)

type Table struct {
	TableName string

	Columns []*Column
}

func (dt *Table) importSql() bool {
	for _, col := range dt.Columns {
		if strings.HasPrefix(col.structFieldType(), "sql.") {
			return true
		}
	}

	return false
}

func (dt *Table) structName() string {

	s := strings.Title(dt.TableName)

	if strings.HasSuffix(s, "ies") {
		s = s[:len(s)-3] + "y"
	} else if strings.HasSuffix(s, "sses") {
		s = s[:len(s)-2]
	} else if strings.HasSuffix(s, "ers") {
		s = s[:len(s)-1]
	} else if strings.HasSuffix(s, "s") {
		s = s[:len(s)-1]
	}

	return reUnderscore.ReplaceAllStringFunc(s, func(s string) string {
		return strings.ToUpper(s[1:])
	})
}

func (dt *Table) keyColumn() *Column {

	for _, c := range dt.Columns {
		if c.Key == "PRI" {
			return c
		}
	}

	return dt.Columns[0] // @todo
}

func (dt *Table) Generate(dir string) error {

	var code bytes.Buffer

	code.WriteString(fmt.Sprintf("package main\n\n"))
	code.WriteString(fmt.Sprintf("import (\n"))
	code.WriteString(fmt.Sprintf("\t\"database/sql\"\n"))
	code.WriteString(fmt.Sprintf("\t\"fmt\"\n"))
	code.WriteString(fmt.Sprintf(")\n\n"))
	code.WriteString(fmt.Sprintf("type %s struct {\n", dt.structName()))

	code.WriteString(fmt.Sprintf("\tdb *sql.DB\n"))
	code.WriteString(fmt.Sprintf("\tTable string\n\n"))

	for _, col := range dt.Columns {
		code.WriteString(fmt.Sprintf("\t%s %s `ot:\"%s\"`\n",
			col.structFieldName(),
			col.structFieldType(),
			col.Type))
	}

	code.WriteString(fmt.Sprintf("}\n\n"))

	// new func
	code.WriteString(fmt.Sprintf("func New%s(db *sql.DB) *%s {\n", dt.structName(), dt.structName()))
	code.WriteString(fmt.Sprintf("\treturn &%s{\n", dt.structName()))
	code.WriteString(fmt.Sprintf("\t\tdb: db,\n"))
	code.WriteString(fmt.Sprintf("\t\tTable: \"%s\",\n", dt.TableName))
	code.WriteString(fmt.Sprintf("\t}\n"))
	code.WriteString(fmt.Sprintf("}\n"))

	// find func
	code.WriteString(fmt.Sprintf("func Find%s(db *sql.DB, key interface{}) *%s {\n", dt.structName(), dt.structName()))
	code.WriteString(fmt.Sprintf("\tobj := New%s(db)\n", dt.structName()))
	code.WriteString(fmt.Sprintf("\trow := obj.db.QueryRow(fmt.Sprintf(\"SELECT * FROM %%s WHERE %s = ?\", obj.Table), key)\n", dt.keyColumn().Name))
	code.WriteString(fmt.Sprintf("\tif row.Err() != nil {\n"))
	code.WriteString(fmt.Sprintf("\t\treturn nil\n"))
	code.WriteString(fmt.Sprintf("\t}\n"))
	code.WriteString(fmt.Sprintf("\trow.Scan(%s)\n", func() string {
		elems := make([]string, len(dt.Columns))

		for i, col := range dt.Columns {
			elems[i] = fmt.Sprintf("&obj.%s", col.structFieldName())
		}

		return strings.Join(elems, ", ")
	}()))
	code.WriteString(fmt.Sprintf("\treturn obj\n"))
	code.WriteString(fmt.Sprintf("}\n"))

	file, err := os.Create(fmt.Sprintf("%s/%s.go", dir, dt.structName()))

	if err != nil {
		return err
	}

	if _, err := file.WriteString(code.String()); err != nil {
		return err
	}

	return nil
}
