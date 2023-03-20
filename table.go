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
	Columns   []*Column
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
	// code.WriteString(fmt.Sprintf("\t\"time\"\n"))
	code.WriteString(fmt.Sprintf("\t\"log\"\n"))
	code.WriteString(fmt.Sprintf(")\n\n"))
	code.WriteString(fmt.Sprintf("type %s struct {\n", dt.structName()))

	code.WriteString(fmt.Sprintf("\tdb *sql.DB\n"))
	code.WriteString(fmt.Sprintf("\tTable string\n\n"))

	for _, col := range dt.Columns {
		code.WriteString(fmt.Sprintf("\t%s %s `dbt:\"%s\"`\n",
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
	code.WriteString(fmt.Sprintf("}\n\n"))

	// find func
	key := dt.keyColumn()

	code.WriteString(fmt.Sprintf("func Find%s(db *sql.DB, %s %s) *%s {\n", dt.structName(), key.structFieldName(), key.structFieldType(), dt.structName()))
	code.WriteString(fmt.Sprintf("\tobj := New%s(db)\n", dt.structName()))
	code.WriteString(fmt.Sprintf("\trow := obj.db.QueryRow(fmt.Sprintf(\"SELECT * FROM %%s WHERE %s = ?\", obj.Table), %s)\n", key.Name, key.structFieldName()))
	code.WriteString(fmt.Sprintf("\tif row.Err() != nil {\n"))
	code.WriteString(fmt.Sprintf("\t\treturn nil\n"))
	code.WriteString(fmt.Sprintf("\t}\n"))
	code.WriteString(fmt.Sprintf("\tif err := row.Scan(%s); err != nil { \n", func() string {
		elems := make([]string, len(dt.Columns))

		for i, col := range dt.Columns {
			elems[i] = fmt.Sprintf("&obj.%s", col.structFieldName())
		}

		return strings.Join(elems, ", ")
	}()))
	code.WriteString(fmt.Sprintf("\t\tlog.Println(err)\n"))
	code.WriteString(fmt.Sprintf("\t\treturn nil\n"))
	code.WriteString(fmt.Sprintf("\t}\n"))
	code.WriteString(fmt.Sprintf("\treturn obj\n"))
	code.WriteString(fmt.Sprintf("}\n"))

	file, err := os.Create(fmt.Sprintf("%s/%s.go", dir, dt.structName()))

	if err != nil {
		return err
	}

	if _, err := file.WriteString(code.String()); err != nil {
		return err
	}

	dt.GenerateTest(dir)

	return nil
}

func (dt *Table) GenerateTest(dir string) error {
	var code bytes.Buffer

	key := dt.keyColumn()

	code.WriteString(fmt.Sprintf("package main\n\n"))
	code.WriteString(fmt.Sprintf("import (\n"))
	code.WriteString(fmt.Sprintf("\t\"fmt\"\n"))
	code.WriteString(fmt.Sprintf("\t\"log\"\n"))
	code.WriteString(fmt.Sprintf("\t\"database/sql\"\n"))
	code.WriteString(fmt.Sprintf("\t_ \"github.com/go-sql-driver/mysql\"\n"))
	code.WriteString(fmt.Sprintf("\t\"testing\"\n"))
	code.WriteString(fmt.Sprintf(")\n\n"))

	code.WriteString(fmt.Sprintf(`func Test%s(t *testing.T) {
	db, err := sql.Open("mysql", "admin:admin@/app")

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	row := db.QueryRow(fmt.Sprintf("select %s from %s limit 1"))

	if row.Err() != nil {
		t.Fatal(row.Err())
		return
	}

	var %s %s

	if err := row.Scan(&%s); err != nil {
		t.Fatal(err)
	}

	obj := Find%s(db, %s)

	if obj == nil {
		t.Fatalf("could not load row %%#v using Find%s", %s)
	} 

	log.Println(obj)
}`, dt.structName(), key.Name, dt.TableName, key.structFieldName(), key.structFieldType(), key.structFieldName(), dt.structName(), key.structFieldName(),
		dt.structName(), key.structFieldName()))

	file, err := os.Create(fmt.Sprintf("%s/%s_test.go", dir, dt.structName()))

	if err != nil {
		return err
	}

	if _, err := file.WriteString(code.String()); err != nil {
		return err
	}

	return nil
}
