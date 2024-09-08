package imports

import (
	"fmt"
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"sort"
)

type Import struct {
	Path  string
	Alias string
}

func (i Import) Format() string {
	if i.Alias == "" {
		return fmt.Sprintf(`"%s"`, i.Path)
	}
	return fmt.Sprintf(`%s "%s"`, i.Alias, i.Path)
}

type ImportBuilder struct {
	imports []Import
	driver  opts.SQLDriver
}

func NewImportBuilder(
	options *opts.Options,
) *ImportBuilder {
	driver := options.Driver()
	return &ImportBuilder{
		driver: driver,
	}
}

func (i *ImportBuilder) AddSqlDriver() *ImportBuilder {
	c := i.clone()
	switch c.driver {
	case opts.SQLDriverPGXV4:
		c.imports = append(c.imports, Import{Path: "github.com/jackc/pgx/v4"})
	case opts.SQLDriverPGXV5:
		c.imports = append(c.imports, Import{Path: "github.com/jackc/pgx/v5"})
	case opts.SQLDriverGoSQLDriverMySQL:
		c.imports = append(c.imports, Import{Path: "github.com/go-sql-driver/mysql"})
	default:
		c.imports = append(c.imports, Import{Path: "database/sql"})
	}

	return c
}

func (i *ImportBuilder) AddConnection() *ImportBuilder {
	c := i.clone()
	switch c.driver {
	case opts.SQLDriverPGXV4:
		c.imports = append(c.imports, Import{Path: "github.com/jackc/pgconn"})
	case opts.SQLDriverPGXV5:
		c.imports = append(c.imports, Import{Path: "github.com/jackc/pgx/v5/pgconn"})
	}

	return c
}

func (i *ImportBuilder) Build() []Import {
	// leave unique only
	unique := make(map[string]Import)
	for _, imp := range i.imports {
		unique[imp.Path] = imp
	}
	i.imports = []Import{}
	for _, imp := range unique {
		i.imports = append(i.imports, imp)
	}
	// sort
	sort.Slice(
		i.imports, func(prev, next int) bool {
			return i.imports[prev].Path < i.imports[next].Path
		},
	)
	return i.imports
}

func (i *ImportBuilder) Clear() *ImportBuilder {
	i.imports = []Import{}
	return i
}

func (i *ImportBuilder) clone() *ImportBuilder {
	c := ImportBuilder{
		imports: i.imports,
		driver:  i.driver,
	}
	return &c
}

func (i *ImportBuilder) Add(importItem Import) *ImportBuilder {
	c := i.clone()
	c.imports = append(c.imports, importItem)
	return c
}

func (i *ImportBuilder) AddWithoutAlias(path string) *ImportBuilder {
	return i.Add(Import{Path: path})
}

func (i *ImportBuilder) AddWithAlias(path, alias string) *ImportBuilder {
	return i.Add(Import{Path: path, Alias: alias})
}
