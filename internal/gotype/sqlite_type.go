package gotype

import (
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"github.com/debugger84/sqlc-dataloader/internal/sqltype"
	"log"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

type SqlLiteTypeTransformer struct {
	customTypes         []sqltype.CustomType
	emitPointersForNull bool
}

func NewSqlLiteTypeTransformer(options *opts.Options, customTypes []sqltype.CustomType) *SqlLiteTypeTransformer {
	emitPointersForNull := options.EmitPointersForNullTypes
	return &SqlLiteTypeTransformer{
		customTypes:         customTypes,
		emitPointersForNull: emitPointersForNull,
	}
}

func (t *SqlLiteTypeTransformer) ToGoType(col *plugin.Column) string {
	dt := strings.ToLower(sdk.DataType(col.Type))
	notNull := col.NotNull || col.IsArray
	emitPointersForNull := t.emitPointersForNull

	switch dt {

	case "int", "integer", "tinyint", "smallint", "mediumint", "bigint", "unsignedbigint", "int2", "int8":
		if notNull {
			return "int64"
		}
		if emitPointersForNull {
			return "*int64"
		}
		return "sql.NullInt64"

	case "blob":
		return "[]byte"

	case "real", "double", "doubleprecision", "float":
		if notNull {
			return "float64"
		}
		if emitPointersForNull {
			return "*float64"
		}
		return "sql.NullFloat64"

	case "boolean", "bool":
		if notNull {
			return "bool"
		}
		if emitPointersForNull {
			return "*bool"
		}
		return "sql.NullBool"

	case "date", "datetime", "timestamp":
		if notNull {
			return "time.Time"
		}
		if emitPointersForNull {
			return "*time.Time"
		}
		return "sql.NullTime"

	case "any":
		return "interface{}"

	}

	switch {

	case strings.HasPrefix(dt, "character"),
		strings.HasPrefix(dt, "varchar"),
		strings.HasPrefix(dt, "varyingcharacter"),
		strings.HasPrefix(dt, "nchar"),
		strings.HasPrefix(dt, "nativecharacter"),
		strings.HasPrefix(dt, "nvarchar"),
		dt == "text",
		dt == "clob":
		if notNull {
			return "string"
		}
		if emitPointersForNull {
			return "*string"
		}
		return "sql.NullString"

	case strings.HasPrefix(dt, "decimal"), dt == "numeric":
		if notNull {
			return "float64"
		}
		if emitPointersForNull {
			return "*float64"
		}
		return "sql.NullFloat64"

	default:
		log.Printf("unknown SQLite type: %s\n", dt)

		return "interface{}"

	}
}
