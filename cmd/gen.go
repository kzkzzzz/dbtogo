package cmd

import (
	"bytes"
	"fmt"
	"github.com/kzkzzzz/dbtogo/common"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
)

type (
	Model struct {
		Name     string
		Table    string
		Receiver string
		Columns  []ColumnInfo
	}
	ColumnInfo struct {
		Table   string `json:"table"`
		Name    string
		Comment string
		Type    string
		GoName  string
		GoType  string
	}
	Gen interface {
		GetColumns() []ColumnInfo
	}
)

func Run(gen Gen) {
	columns := gen.GetColumns()
	tc := make(map[string][]ColumnInfo, 0)
	for _, column := range columns {
		if column.Comment == "" {
			column.Comment = column.GoName
		}
		tc[column.Table] = append(tc[column.Table], column)
	}

	tmpl, err := template.ParseFiles("model.tmpl")
	if err != nil {
		common.Log.Errorf("加载模板失败: %s", err)
		return
	}

	line := strings.Repeat("-", 16)

	for _, table := range cmdParam.Table {
		tColumns, ok := tc[table]
		if !ok {
			continue
		}

		name := common.StrToCamelCase(table)
		m := Model{
			Name:     name,
			Table:    table,
			Receiver: fmt.Sprintf("%s *%s", strings.ToLower(name[:1]), name),
			Columns:  tColumns,
		}

		buf := bytes.NewBuffer(nil)
		err = tmpl.Execute(buf, m)
		if err != nil {
			common.Log.Errorf("渲染模板失败: %s", err)
			continue
		}

		source, err := format.Source(buf.Bytes())
		if err != nil {
			common.Log.Errorf("格式化模板代码失败: %s", err)
			continue
		}

		tLine := fmt.Sprintf("%s %s %s", line, table, line)
		fmt.Printf("\n%s\n%s\n%s\n\n", tLine, string(source), tLine)

		if cmdParam.Output != "" {
			filename := filepath.Join(cmdParam.Output, fmt.Sprintf("%s.go", table))
			err = ioutil.WriteFile(filename, source, 0644)
			if err != nil {
				common.Log.Errorf("写入文件失败: %s", err)
			} else {
				common.Log.Infof("写入文件%s", filename)
			}
		}

	}
}
