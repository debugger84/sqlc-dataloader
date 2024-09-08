package sqltype

import (
	"github.com/debugger84/sqlc-dataloader/internal/naming"
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type CustomTypeKind string

const (
	EnumType      CustomTypeKind = "enum"
	CompositeType CustomTypeKind = "composite"
)

type CustomType struct {
	GoTypeName  string
	SqlTypeName string
	Schema      string
	Kind        CustomTypeKind
	IsNullable  bool
}

func NewCustomTypes(
	schemas []*plugin.Schema,
	options *opts.Options,
) []CustomType {
	normalizer := naming.NewNameNormalizer(options)
	driver := options.Driver()
	customTypes := make([]CustomType, 0)
	for _, schema := range schemas {
		if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
			continue
		}

		for _, enum := range schema.Enums {
			baseName := normalizer.NormalizeSqlName(schema.Name, enum.Name)
			customTypes = append(
				customTypes, CustomType{
					GoTypeName:  "Null" + normalizer.NormalizeGoType(baseName),
					SqlTypeName: enum.Name,
					Schema:      schema.Name,
					Kind:        EnumType,
					IsNullable:  true,
				}, CustomType{
					GoTypeName:  normalizer.NormalizeGoType(baseName),
					SqlTypeName: enum.Name,
					Schema:      schema.Name,
					Kind:        EnumType,
					IsNullable:  false,
				},
			)
		}

		emitPointersForNull := driver.IsPGX() && options.EmitPointersForNullTypes

		for _, ct := range schema.CompositeTypes {
			name := "string"
			nullName := "sql.NullString"
			if emitPointersForNull {
				nullName = "*string"
			}

			customTypes = append(
				customTypes, CustomType{
					GoTypeName:  name,
					SqlTypeName: ct.Name,
					Schema:      schema.Name,
					Kind:        CompositeType,
					IsNullable:  false,
				}, CustomType{
					GoTypeName:  nullName,
					SqlTypeName: ct.Name,
					Schema:      schema.Name,
					Kind:        CompositeType,
					IsNullable:  true,
				},
			)
		}
	}

	return customTypes
}
