package model

import "github.com/sqlc-dev/plugin-sdk-go/plugin"

type Field struct {
	name     string // CamelCased name for Go
	dBName   string // Name as used in the DB
	typeName string
	tags     map[string]string
	comment  string
	column   *plugin.Column

	isPrimaryKey bool

	// EmbedFields contains the embedded fields that require scanning.
	embedFields []Field
}

func (f *Field) Name() string {
	return f.name
}

func (f *Field) DBName() string {
	return f.dBName
}

func (f *Field) Type() string {
	return f.typeName
}

func (f *Field) Tags() map[string]string {
	return f.tags
}

func (f *Field) Comment() string {
	return f.comment
}

func (f *Field) Column() *plugin.Column {
	return f.column
}

func (f *Field) EmbedFields() []Field {
	return f.embedFields
}

func (f *Field) IsPrimaryKey() bool {
	return f.isPrimaryKey
}
