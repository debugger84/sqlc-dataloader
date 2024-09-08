package gotype

import (
	"fmt"
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"github.com/debugger84/sqlc-dataloader/internal/sqltype"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

type DbTOGoTypeTransformer interface {
	ToGoType(col *plugin.Column) string
}

type GoTypeFormatter struct {
	defaultSchema      string
	sqlTypeTransformer DbTOGoTypeTransformer
	options            *opts.Options
}

func NewGoTypeFormatter(
	typeTransformer DbTOGoTypeTransformer,
	options *opts.Options,
) *GoTypeFormatter {
	defaultSchema := options.DefaultSchema
	return &GoTypeFormatter{
		defaultSchema:      defaultSchema,
		sqlTypeTransformer: typeTransformer,
		options:            options,
	}
}

func NewDbTOGoTypeTransformer(
	engine opts.SQLEngine,
	customTypes []sqltype.CustomType,
	options *opts.Options,
) (DbTOGoTypeTransformer, error) {
	var typeTransformer DbTOGoTypeTransformer
	switch engine {
	case opts.SQLEngineMySQL:
		return NewMysqlTypeTransformer(customTypes), nil
	case opts.SQLEngineSQLite:
		return NewSqlLiteTypeTransformer(options, customTypes), nil
	case opts.SQLEnginePostgresql:
		return NewPostgresqlTypeTransformer(options, customTypes), nil
	}
	return typeTransformer, fmt.Errorf("unsupported sql engine %s", engine)
}

func (f *GoTypeFormatter) ToGoType(col *plugin.Column) string {
	gotype := f.overriddenType(col)
	if gotype == "" {
		gotype = f.sqlTypeTransformer.ToGoType(col)
	}
	if col.IsSqlcSlice {
		return "[]" + gotype
	}
	if col.IsArray {
		return strings.Repeat("[]", int(col.ArrayDims)) + gotype
	}
	return gotype
}

func (f *GoTypeFormatter) overriddenType(col *plugin.Column) string {
	columnType := sdk.DataType(col.Type)
	notNull := col.NotNull || col.IsArray

	// Check if the column's type has been overridden
	for _, override := range f.options.Overrides {
		oride := override.ShimOverride

		if oride.GoType.TypeName == "" {
			continue
		}
		cname := col.Name
		if col.OriginalName != "" {
			cname = col.OriginalName
		}
		sameTable := override.Matches(col.Table, f.defaultSchema)
		if oride.Column != "" && sdk.MatchString(oride.ColumnName, cname) && sameTable {
			return oride.GoType.TypeName
		}
		if oride.DbType != "" && oride.DbType == columnType && oride.Nullable != notNull && oride.Unsigned == col.Unsigned {
			return oride.GoType.TypeName
		}
	}

	return ""
}
