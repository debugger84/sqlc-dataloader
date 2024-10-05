package internal

import (
	"context"
	"github.com/debugger84/sqlc-dataloader/internal/imports"
	"github.com/debugger84/sqlc-dataloader/internal/model"
	"github.com/debugger84/sqlc-dataloader/internal/opts"
	"github.com/debugger84/sqlc-dataloader/internal/renderer"
	"github.com/debugger84/sqlc-dataloader/internal/sqltype"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
	"strings"
)

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	options, err := opts.Parse(req)
	if err != nil {
		return nil, err
	}

	if err := opts.ValidateOpts(options); err != nil {
		return nil, err
	}

	if options.DefaultSchema != "" {
		req.Catalog.DefaultSchema = options.DefaultSchema
	}
	modelPkg := ""
	if options.ModelImport != "" {
		pkgParts := strings.Split(options.ModelImport, "/")
		modelPkg = pkgParts[len(pkgParts)-1]
	}
	customTypes := sqltype.NewCustomTypes(req.Catalog.Schemas, options, modelPkg)
	structs := model.BuildStructs(req, options, customTypes)

	importer := imports.NewImportBuilder(options)

	loaderRendered := renderer.NewDataLoaderRenderer(structs, options, importer)

	files := make([]*plugin.File, 0)
	loaderFiles, err := loaderRendered.Render()
	if err != nil {
		return nil, err
	}
	files = append(files, loaderFiles...)

	return &plugin.GenerateResponse{
		Files: files,
	}, nil
}
