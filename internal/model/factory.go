package model

import (
	"github.com/debugger84/sqlc-dataloader/internal/gotype"
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"github.com/debugger84/sqlc-dataloader/internal/sqltype"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"log"
	"sort"
)

func BuildStructs(
	req *plugin.GenerateRequest,
	options *opts.Options,
	customTypes []sqltype.CustomType,
) []Struct {
	var structs []Struct

	gotypeTransformer, err := gotype.NewDbTOGoTypeTransformer(opts.SQLEngine(req.Settings.Engine), customTypes, options)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	goTypeFormatter := gotype.NewGoTypeFormatter(gotypeTransformer, options)
	for _, schema := range req.Catalog.Schemas {
		if schema.Name == "pg_catalog" || schema.Name == "information_schema" {
			continue
		}
		for _, table := range schema.Tables {
			s := NewStruct(table, options, goTypeFormatter)
			structs = append(structs, *s)
		}
	}
	if len(structs) > 0 {
		sort.Slice(structs, func(i, j int) bool { return structs[i].Type().TypeName() < structs[j].Type().TypeName() })
	}
	return structs
}
