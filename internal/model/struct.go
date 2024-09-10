package model

import (
	"fmt"
	gotype "github.com/debugger84/sqlc-dataloader/internal/gotype"
	"github.com/debugger84/sqlc-dataloader/internal/imports"
	"github.com/debugger84/sqlc-dataloader/internal/inflection"
	"github.com/debugger84/sqlc-dataloader/internal/naming"
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"strings"
)

type Struct struct {
	table         *plugin.Table
	tableName     string
	structName    string
	fields        []Field
	hasPrimaryKey bool
	goType        *gotype.GoType
}

func NewStruct(
	table *plugin.Table,
	options *opts.Options,
	goTypeFormatter *gotype.GoTypeFormatter,
) *Struct {
	nameNormalizer := naming.NewNameNormalizer(options)
	s := &Struct{
		table: table,
	}

	s.initNames(table, options, nameNormalizer)
	s.initFields(table, options, nameNormalizer, goTypeFormatter)

	return s
}

func (s *Struct) initNames(
	table *plugin.Table,
	options *opts.Options,
	normalizer *naming.NameNormalizer,
) {
	schema := table.Rel.GetSchema()
	tableName := table.Rel.GetName()
	s.tableName = normalizer.NormalizeSqlName(schema, tableName)

	structName := s.tableName
	if !options.EmitExactTableNames {
		structName = inflection.Singular(
			inflection.SingularParams{
				Name:       structName,
				Exclusions: options.InflectionExcludeTableNames,
			},
		)
	}

	structName = normalizer.NormalizeGoType(structName)
	if options.ModelImport != "" {
		structName = options.ModelImport + "." + structName
	}
	s.goType = gotype.NewGoType(structName)
}

func (s *Struct) TableName() string {
	return s.tableName
}

func (s *Struct) FullTableName() string {
	schema := s.table.Rel.GetSchema()
	tableName := s.table.Rel.GetName()

	if schema == "" {
		return tableName
	}

	return fmt.Sprintf("%s.%s", schema, tableName)
}

func (s *Struct) Type() *gotype.GoType {
	return s.goType
}

func (s *Struct) GetImports() []imports.Import {
	if s.goType == nil {
		return nil
	}
	if s.goType.Import().Path == "" {
		return nil
	}
	return []imports.Import{
		s.goType.Import(),
	}
}

func (s *Struct) initFields(
	table *plugin.Table,
	options *opts.Options,
	normalizer *naming.NameNormalizer,
	goTypeFormatter *gotype.GoTypeFormatter,
) {
	primaryKeyColumn := "id"
	for _, column := range options.PrimaryKeysColumns {
		parts := strings.Split(column, ".")
		if len(parts) == 2 && parts[0] == table.Rel.GetName() {
			primaryKeyColumn = parts[1]
			break
		}
	}
	for _, column := range table.Columns {
		tags := map[string]string{}
		isPrimaryKey := false
		if column.Name == primaryKeyColumn {
			isPrimaryKey = true
			s.hasPrimaryKey = true
		}
		goType := goTypeFormatter.ToGoType(column)
		s.fields = append(
			s.fields, Field{
				name:         normalizer.NormalizeGoType(column.Name),
				dBName:       column.Name,
				goType:       &goType,
				tags:         tags,
				comment:      column.Comment,
				column:       column,
				isPrimaryKey: isPrimaryKey,
			},
		)
	}
}

func (s *Struct) Fields() []Field {
	return s.fields
}

func (s *Struct) HasPrimaryKey() bool {
	return s.hasPrimaryKey
}
