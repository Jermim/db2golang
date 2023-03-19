package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var reNumberBetweenBrackets = regexp.MustCompile("(?si)(\\(\\d+\\))")

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

func (dt *Table) Generate(dir string) error {

	var code bytes.Buffer

	code.WriteString(fmt.Sprintf("package main\n\n"))

	if dt.importSql() {
		code.WriteString(fmt.Sprintf("import \"database/sql\"\n\n"))
	}
	code.WriteString(fmt.Sprintf("type %s struct {\n", dt.TableName))

	for _, col := range dt.Columns {
		code.WriteString(fmt.Sprintf("\t%s %s `ot:\"%s\"`\n",
			col.structFieldName(),
			col.structFieldType(),
			col.Type))
	}

	code.WriteString(fmt.Sprintf("}"))

	file, err := os.Create(fmt.Sprintf("%s/%s.go", dir, dt.TableName))

	if err != nil {
		return err
	}

	if _, err := file.WriteString(code.String()); err != nil {
		return err
	}

	return nil
}
