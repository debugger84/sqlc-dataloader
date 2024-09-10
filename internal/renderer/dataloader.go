package renderer

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/debugger84/sqlc-dataloader/internal/imports"
	"github.com/debugger84/sqlc-dataloader/internal/model"
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"github.com/iancoleman/strcase"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
	"go/format"
	"text/template"
)

type DataLoaderTplData struct {
	Struct               model.Struct
	Package              string
	PrimaryKeyColumnName string
	PrimaryKeyFieldType  string
	PrimaryKeyFieldName  string
	Imports              []imports.Import
}

type DataLoaderRenderer struct {
	structs       []model.Struct
	loaderPackage string
	importer      *imports.ImportBuilder
}

func NewDataLoaderRenderer(
	structs []model.Struct,
	options *opts.Options,
	importer *imports.ImportBuilder,
) *DataLoaderRenderer {
	return &DataLoaderRenderer{
		structs:       structs,
		loaderPackage: options.Package,
		importer:      importer,
	}
}

func (r *DataLoaderRenderer) Render() ([]*plugin.File, error) {
	funcMap := template.FuncMap{
		"lowerTitle": sdk.LowerTitle,
	}
	tmpl := template.Must(
		template.New("dataloader.tmpl").
			Funcs(funcMap).
			ParseFS(
				templates,
				"templates/dataloader.tmpl",
			),
	)
	files := make([]*plugin.File, 0)
	importer := r.importer.
		AddSqlDriver().
		AddWithoutAlias("context").
		AddWithoutAlias("github.com/graph-gophers/dataloader/v7")
	for _, s := range r.structs {
		if !s.HasPrimaryKey() {
			continue
		}
		file, err := r.renderDataLoader(tmpl, s, importer)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func (r *DataLoaderRenderer) renderDataLoader(
	tmpl *template.Template,
	s model.Struct,
	importer *imports.ImportBuilder,
) (*plugin.File, error) {
	var pkField model.Field
	for _, f := range s.Fields() {
		if f.IsPrimaryKey() {
			pkField = f
			break
		}
	}

	tctx := DataLoaderTplData{
		Struct:               s,
		Package:              r.loaderPackage,
		PrimaryKeyColumnName: pkField.DBName(),
		PrimaryKeyFieldType:  pkField.Type().TypeWithPackage(),
		PrimaryKeyFieldName:  pkField.Name(),
		Imports: importer.
			Add(pkField.Type().Import()).
			ImportContainer(&s).
			Build(),
	}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := tmpl.ExecuteTemplate(w, "dataloader.tmpl", &tctx)
	w.Flush()
	if err != nil {
		return nil, err
	}
	code, err := format.Source(b.Bytes())
	if err != nil {
		fmt.Println(b.String())
		return nil, fmt.Errorf("source error: %w", err)
	}
	filename := fmt.Sprintf("%s_loader.go", strcase.ToSnake(s.Type().TypeName()))
	if r.loaderPackage != s.Type().PackageName() {
		filename = fmt.Sprintf("%s/%s.go", r.loaderPackage, strcase.ToSnake(s.Type().TypeName()))
	}
	file := &plugin.File{
		Name:     filename,
		Contents: code,
	}
	return file, nil
}
