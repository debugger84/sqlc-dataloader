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
	Struct               LoaderStruct
	Package              string
	PrimaryKeyColumnName string
	PrimaryKeyFieldType  string
	PrimaryKeyFieldName  string
	Imports              []imports.Import
}

type LoaderFactoryTplData struct {
	Structs      []LoaderStruct
	Package      string
	Imports      []imports.Import
	ModelPackage string
}

type DataLoaderRenderer struct {
	structs       []LoaderStruct
	loaderPackage string
	importer      *imports.ImportBuilder
}

type LoaderStruct struct {
	model.Struct
	LoaderName string
	Cache      opts.Cache
}

func NewDataLoaderRenderer(
	structs []model.Struct,
	options *opts.Options,
	importer *imports.ImportBuilder,
) *DataLoaderRenderer {
	loaderStructs := make([]LoaderStruct, 0, len(structs))
	defCache := opts.Cache{
		Type: "no-cache",
	}
	for _, s := range structs {
		structCache := defCache
		loaderName := fmt.Sprintf("%sLoader", s.Type().TypeName())
		for _, cache := range options.Cache {
			if cache.Table == s.FullTableName() &&
				(cache.Type == "lru" || cache.Type == "memory") {
				structCache = cache
				break
			}
		}

		skip := false
		for _, table := range options.ExcludeTables {
			if table == s.FullTableName() {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		loaderStructs = append(
			loaderStructs, LoaderStruct{
				Struct:     s,
				LoaderName: loaderName,
				Cache:      structCache,
			},
		)
	}

	return &DataLoaderRenderer{
		structs:       loaderStructs,
		loaderPackage: options.Package,
		importer:      importer,
	}
}

func (r *DataLoaderRenderer) Render() ([]*plugin.File, error) {
	if len(r.structs) == 0 {
		return nil, nil
	}
	funcMap := template.FuncMap{
		"lowerTitle": sdk.LowerTitle,
	}
	tmpl := template.Must(
		template.New("dataloader.tmpl").
			Funcs(funcMap).
			ParseFS(
				templates,
				"templates/dataloader.tmpl",
				"templates/loader_factory.tmpl",
			),
	)
	files := make([]*plugin.File, 0)
	loaderImporter := r.importer.
		AddSqlDriver().
		AddWithoutAlias("context").
		AddWithoutAlias("github.com/graph-gophers/dataloader/v7")

	for _, s := range r.structs {
		if !s.HasPrimaryKey() {
			continue
		}
		file, err := r.renderDataLoader(tmpl, s, loaderImporter)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	factoryImporter := r.importer
	file, err := r.renderLoaderFactory(tmpl, r.structs, factoryImporter)
	if err != nil {
		return nil, err
	}
	files = append(files, file)

	return files, nil
}

func (r *DataLoaderRenderer) renderLoaderFactory(
	tmpl *template.Template,
	structs []LoaderStruct,
	importer *imports.ImportBuilder,
) (*plugin.File, error) {
	s := structs[0]
	tctx := LoaderFactoryTplData{
		Structs:      structs,
		Package:      r.loaderPackage,
		ModelPackage: s.Type().PackageName(),
		Imports: importer.
			ImportContainer(&s).
			Build(),
	}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	err := tmpl.ExecuteTemplate(w, "loader_factory.tmpl", &tctx)
	w.Flush()
	if err != nil {
		return nil, err
	}
	code, err := format.Source(b.Bytes())
	if err != nil {
		fmt.Println(b.String())
		return nil, fmt.Errorf("source error: %w", err)
	}
	filename := fmt.Sprintf("loader_factory.go")
	if r.loaderPackage != s.Type().PackageName() {
		filename = fmt.Sprintf("%s/%s", r.loaderPackage, filename)
	}

	file := &plugin.File{
		Name:     filename,
		Contents: code,
	}
	return file, nil
}

func (r *DataLoaderRenderer) renderDataLoader(
	tmpl *template.Template,
	s LoaderStruct,
	importer *imports.ImportBuilder,
) (*plugin.File, error) {
	var pkField model.Field
	for _, f := range s.Fields() {
		if f.IsPrimaryKey() {
			pkField = f
			break
		}
	}

	if s.Cache.Type == "lru" {
		importer = importer.AddWithAlias("github.com/debugger84/sqlc-dataloader/cache", "loaderCache").
			AddWithoutAlias("time")
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
