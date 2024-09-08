package naming

import (
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"strings"
	"unicode"
	"unicode/utf8"
)

type NameNormalizer struct {
	options *opts.Options
}

func NewNameNormalizer(options *opts.Options) *NameNormalizer {
	return &NameNormalizer{
		options: options,
	}
}

func (n *NameNormalizer) NormalizeGoType(name string) string {
	out := ""
	if rename := n.options.Rename[name]; rename != "" {
		return rename
	}
	name = strings.Map(
		func(r rune) rune {
			if unicode.IsLetter(r) {
				return r
			}
			if unicode.IsDigit(r) {
				return r
			}
			return rune('_')
		}, name,
	)

	for _, p := range strings.Split(name, "_") {
		if _, found := n.options.InitialismsMap[p]; found {
			out += strings.ToUpper(p)
		} else {
			out += strings.Title(p)
		}
	}

	// If a name has a digit as its first char, prepand an underscore to make it a valid Go name.
	r, _ := utf8.DecodeRuneInString(out)
	if unicode.IsDigit(r) {
		return "_" + out
	} else {
		return out
	}
}

func (n *NameNormalizer) NormalizeSqlName(schema, tableName string) string {
	defSchema := n.options.DefaultSchema
	if schema != defSchema {
		tableName = schema + "_" + tableName
	}
	return tableName
}

func (n *NameNormalizer) NormalizeCompositeTypeName(schema, tableName string) string {
	defSchema := n.options.DefaultSchema
	if schema != defSchema {
		tableName = schema + "_" + tableName
	}
	return tableName
}
