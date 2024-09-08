package gotype

import (
	"github.com/debugger84/sqlc-dataloader/internal/sqltype"
	"log"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"github.com/sqlc-dev/plugin-sdk-go/sdk"
)

type MysqlTypeTransformer struct {
	customTypes []sqltype.CustomType
}

func NewMysqlTypeTransformer(customTypes []sqltype.CustomType) *MysqlTypeTransformer {
	return &MysqlTypeTransformer{
		customTypes: customTypes,
	}
}

func (t *MysqlTypeTransformer) ToGoType(col *plugin.Column) string {
	columnType := sdk.DataType(col.Type)
	notNull := col.NotNull || col.IsArray
	unsigned := col.Unsigned

	switch columnType {

	case "varchar", "text", "char", "tinytext", "mediumtext", "longtext":
		if notNull {
			return "string"
		}
		return "sql.NullString"

	case "tinyint":
		if col.Length == 1 {
			if notNull {
				return "bool"
			}
			return "sql.NullBool"
		} else {
			if notNull {
				if unsigned {
					return "uint8"
				}
				return "int8"
			}
			// The database/sql package does not have a sql.NullInt8 type, so we
			// use the smallest type they have which is NullInt16
			return "sql.NullInt16"
		}

	case "year":
		if notNull {
			return "int16"
		}
		return "sql.NullInt16"

	case "smallint":
		if notNull {
			if unsigned {
				return "uint16"
			}
			return "int16"
		}
		return "sql.NullInt16"

	case "int", "integer", "mediumint":
		if notNull {
			if unsigned {
				return "uint32"
			}
			return "int32"
		}
		return "sql.NullInt32"

	case "bigint":
		if notNull {
			if unsigned {
				return "uint64"
			}
			return "int64"
		}
		return "sql.NullInt64"

	case "blob", "binary", "varbinary", "tinyblob", "mediumblob", "longblob":
		if notNull {
			return "[]byte"
		}
		return "sql.NullString"

	case "double", "double precision", "real", "float":
		if notNull {
			return "float64"
		}
		return "sql.NullFloat64"

	case "decimal", "dec", "fixed":
		if notNull {
			return "string"
		}
		return "sql.NullString"

	case "enum":
		// TODO: Proper Enum support
		return "string"

	case "date", "timestamp", "datetime", "time":
		if notNull {
			return "time.Time"
		}
		return "sql.NullTime"

	case "boolean", "bool":
		if notNull {
			return "bool"
		}
		return "sql.NullBool"

	case "json":
		return "json.RawMessage"

	case "any":
		return "interface{}"

	default:
		for _, customType := range t.customTypes {
			if col.Type.Name == customType.SqlTypeName &&
				notNull == !customType.IsNullable {
				return customType.GoTypeName
			}
		}
		log.Printf("Unknown MySQL type: %s\n", columnType)
		return "interface{}"

	}
}
