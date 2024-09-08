package opts

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Options struct {
	EmitExactTableNames         bool              `json:"emit_exact_table_names,omitempty" yaml:"emit_exact_table_names"`
	Package                     string            `json:"package" yaml:"package"`
	Out                         string            `json:"out" yaml:"out"`
	Overrides                   []Override        `json:"overrides,omitempty" yaml:"overrides"`
	Rename                      map[string]string `json:"rename,omitempty" yaml:"rename"`
	OmitSqlcVersion             bool              `json:"omit_sqlc_version,omitempty" yaml:"omit_sqlc_version"`
	DefaultSchema               string            `json:"default_schema,omitempty" yaml:"default_schema"`
	InflectionExcludeTableNames []string          `json:"inflection_exclude_table_names,omitempty" yaml:"inflection_exclude_table_names"`
	Initialisms                 *[]string         `json:"initialisms,omitempty" yaml:"initialisms"`
	SqlPackage                  string            `json:"sql_package" yaml:"sql_package"`
	EmitPointersForNullTypes    bool              `json:"emit_pointers_for_null_types" yaml:"emit_pointers_for_null_types"`
	PrimaryKeysColumns          []string          `json:"primary_keys_columns" yaml:"primary_keys_columns"`

	InitialismsMap map[string]struct{} `json:"-" yaml:"-"`
}

type GlobalOptions struct {
	Overrides []Override        `json:"overrides,omitempty" yaml:"overrides"`
	Rename    map[string]string `json:"rename,omitempty" yaml:"rename"`
}

func Parse(req *plugin.GenerateRequest) (*Options, error) {
	options, err := parseOpts(req)
	if err != nil {
		return nil, err
	}
	global, err := parseGlobalOpts(req)
	if err != nil {
		return nil, err
	}
	if len(global.Overrides) > 0 {
		options.Overrides = append(global.Overrides, options.Overrides...)
	}
	if len(global.Rename) > 0 {
		if options.Rename == nil {
			options.Rename = map[string]string{}
		}
		maps.Copy(options.Rename, global.Rename)
	}
	return options, nil
}

func parseOpts(req *plugin.GenerateRequest) (*Options, error) {
	var options Options
	if len(req.PluginOptions) == 0 {
		return &options, nil
	}
	if err := json.Unmarshal(req.PluginOptions, &options); err != nil {
		return nil, fmt.Errorf("unmarshalling plugin options: %w", err)
	}

	if options.Package == "" {
		if options.Out != "" {
			options.Package = filepath.Base(options.Out)
		} else {
			return nil, fmt.Errorf("invalid options: missing package name")
		}
	}

	for i := range options.Overrides {
		if err := options.Overrides[i].parse(req); err != nil {
			return nil, err
		}
	}

	if options.Initialisms == nil {
		options.Initialisms = new([]string)
		*options.Initialisms = []string{"id"}
	}

	options.InitialismsMap = map[string]struct{}{}
	for _, initial := range *options.Initialisms {
		options.InitialismsMap[initial] = struct{}{}
	}

	return &options, nil
}

func parseGlobalOpts(req *plugin.GenerateRequest) (*GlobalOptions, error) {
	var options GlobalOptions
	if len(req.GlobalOptions) == 0 {
		return &options, nil
	}
	if err := json.Unmarshal(req.GlobalOptions, &options); err != nil {
		return nil, fmt.Errorf("unmarshalling global options: %w", err)
	}
	for i := range options.Overrides {
		if err := options.Overrides[i].parse(req); err != nil {
			return nil, err
		}
	}
	return &options, nil
}

func ValidateOpts(opts *Options) error {

	return nil
}

func (o *Options) Driver() SQLDriver {
	return NewSQLDriver(o.SqlPackage)
}
