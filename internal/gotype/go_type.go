package gotype

import (
	"fmt"
	"github.com/debugger84/sqlc-dataloader/internal/imports"
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"github.com/debugger84/sqlc-dataloader/internal/sqltype"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

type GoType struct {
	typeName    string
	packageName string
	typeImport  imports.Import
	isPointer   bool
	isArray     bool
	arrayDims   int
}

// NewGoType creates a new GoType from a full type name
// e.g. plugin.Table -> GoType{typeName: "Table", packageName: "plugin"}
// e.g. *plugin.Table -> GoType{typeName: "Table", packageName: "plugin", isPointer: true}
// e.g. []plugin.Table -> GoType{typeName: "Table", packageName: "plugin", isArray: true, arrayDims: 1}
// e.g. github.com/sqlc-dev/plugin-sdk-go/plugin.Table -> GoType{typeName: "Table", packageName: "plugin", typeImport: Import{Path: "github.com/sqlc-dev/plugin-sdk-go/plugin"}}
func NewGoType(fullTypeName string) *GoType {
	isArray := strings.HasPrefix(fullTypeName, "[]")
	arrayDims := 0
	if isArray {
		for strings.HasPrefix(fullTypeName, "[]") {
			fullTypeName = fullTypeName[2:]
			arrayDims++
		}
	}
	isPointer := strings.HasPrefix(fullTypeName, "*")
	if isPointer {
		fullTypeName = fullTypeName[1:]
	}
	parts := strings.Split(fullTypeName, ".")
	if len(parts) == 2 {
		pathParts := strings.Split(parts[0], "/")
		importPath := ""
		packageName := parts[0]
		if len(pathParts) > 1 {
			importPath = parts[0]
			packageName = pathParts[len(pathParts)-1]
		}
		return &GoType{
			typeName:    parts[1],
			packageName: packageName,
			isPointer:   isPointer,
			isArray:     isArray,
			arrayDims:   arrayDims,
			typeImport: imports.Import{
				Path: importPath,
			},
		}
	}
	return &GoType{
		typeName:  fullTypeName,
		isPointer: isPointer,
		isArray:   isArray,
		arrayDims: arrayDims,
	}
}

func (g *GoType) String() string {
	str := g.TypeWithPackage()
	if g.isPointer {
		str = "*" + str
	}
	if g.isArray {
		str = strings.Repeat("[]", g.arrayDims) + str
	}

	return str
}

func (g *GoType) TypeWithPackage() string {
	str := g.typeName

	if g.packageName != "" {
		str = fmt.Sprintf("%s.%s", g.packageName, str)
	}

	return str
}

func (g *GoType) Import() imports.Import {
	return g.typeImport
}

func (g *GoType) IsPointer() bool {
	return g.isPointer
}

func (g *GoType) SetImport(imp imports.Import) *GoType {
	g.typeImport = imp
	return g
}

func (g *GoType) IsArray() bool {
	return g.isArray
}

func (g *GoType) TypeName() string {
	return g.typeName
}

func (g *GoType) PackageName() string {
	return g.packageName
}

type DbTOGoTypeTransformer interface {
	ToGoType(col *plugin.Column) GoType
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

func (f *GoTypeFormatter) ToGoType(col *plugin.Column) GoType {
	gotype, overridden := f.overriddenType(col)
	if !overridden {
		gotype = f.sqlTypeTransformer.ToGoType(col)
	}
	if gotype.packageName != "" && gotype.typeImport.Path == "" {
		gotype = f.addImport(gotype)
	}

	if col.IsSqlcSlice {
		gotype.isArray = true
		if col.IsArray {
			gotype.arrayDims = int(col.ArrayDims)
		}
	}

	return gotype
}

func (f *GoTypeFormatter) addImport(goType GoType) GoType {
	if goType.PackageName() == "" {
		return goType
	}
	switch goType.PackageName() {
	case "sql":
		return *goType.SetImport(
			imports.Import{
				Path: "database/sql",
			},
		)
	case "uuid":
		return *goType.SetImport(
			imports.Import{
				Path: "github.com/google/uuid",
			},
		)
	case "netip":
		return *goType.SetImport(
			imports.Import{
				Path: "net/netip",
			},
		)
	case "time":
		return *goType.SetImport(
			imports.Import{
				Path: "time",
			},
		)
	case "json":
		return *goType.SetImport(
			imports.Import{
				Path: "encoding/json",
			},
		)
	case "net":
		return *goType.SetImport(
			imports.Import{
				Path: "net",
			},
		)

	}

	return goType
}

func (f *GoTypeFormatter) overriddenType(col *plugin.Column) (GoType, bool) {
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
			return *NewGoType(oride.GoType.TypeName).SetImport(
				imports.Import{
					Path:  oride.GoType.ImportPath,
					Alias: oride.GoType.Package,
				},
			), true
		}
		if oride.DbType != "" && oride.DbType == columnType && oride.Nullable != notNull && oride.Unsigned == col.Unsigned {
			return *NewGoType(oride.GoType.TypeName).SetImport(
				imports.Import{
					Path:  oride.GoType.ImportPath,
					Alias: oride.GoType.Package,
				},
			), true
		}
	}

	return GoType{}, false
}
