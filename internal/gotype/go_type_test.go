package gotype

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGoType(t *testing.T) {
	t.Run("simple type without package", func(t *testing.T) {
		gt := NewGoType("string")
		require.NotNil(t, gt)
		require.Equal(t, "string", gt.String())
		require.Equal(t, "string", gt.TypeName())
		require.Equal(t, "", gt.PackageName())
		require.False(t, gt.IsPointer())
		require.False(t, gt.IsArray())
	})

	t.Run("plugin.Table -> GoType{typeName: Table, packageName: plugin}", func(t *testing.T) {
		gt := NewGoType("plugin.Table")
		require.NotNil(t, gt)
		require.Equal(t, "plugin.Table", gt.String())
		require.Equal(t, "Table", gt.TypeName())
		require.Equal(t, "plugin", gt.PackageName())
		require.False(t, gt.IsPointer())
		require.False(t, gt.IsArray())
		require.Equal(t, "", gt.Import().Path)
	})

	t.Run("*plugin.Table -> GoType{typeName: Table, packageName: plugin, isPointer: true}", func(t *testing.T) {
		gt := NewGoType("*plugin.Table")
		require.NotNil(t, gt)
		require.Equal(t, "*plugin.Table", gt.String())
		require.Equal(t, "Table", gt.TypeName())
		require.Equal(t, "plugin", gt.PackageName())
		require.True(t, gt.IsPointer())
		require.False(t, gt.IsArray())
		require.Equal(t, "", gt.Import().Path)
	})

	t.Run("[]plugin.Table -> GoType{typeName: Table, packageName: plugin, isArray: true, arrayDims: 1}", func(t *testing.T) {
		gt := NewGoType("[]plugin.Table")
		require.NotNil(t, gt)
		require.Equal(t, "[]plugin.Table", gt.String())
		require.Equal(t, "Table", gt.TypeName())
		require.Equal(t, "plugin", gt.PackageName())
		require.False(t, gt.IsPointer())
		require.True(t, gt.IsArray())
		require.Equal(t, "", gt.Import().Path)
	})

	t.Run("github.com/sqlc-dev/plugin-sdk-go/plugin.Table -> GoType{typeName: Table, packageName: plugin, typeImport: Import{Path: github.com/sqlc-dev/plugin-sdk-go/plugin}}", func(t *testing.T) {
		gt := NewGoType("github.com/sqlc-dev/plugin-sdk-go/plugin.Table")
		require.NotNil(t, gt)
		require.Equal(t, "plugin.Table", gt.String())
		require.Equal(t, "Table", gt.TypeName())
		require.Equal(t, "plugin", gt.PackageName())
		require.False(t, gt.IsPointer())
		require.False(t, gt.IsArray())
		require.Equal(t, "github.com/sqlc-dev/plugin-sdk-go/plugin", gt.Import().Path)
	})

	t.Run("multidimensional array [][]plugin.Table", func(t *testing.T) {
		gt := NewGoType("[][]plugin.Table")
		require.NotNil(t, gt)
		require.Equal(t, "[][]plugin.Table", gt.String())
		require.Equal(t, "Table", gt.TypeName())
		require.Equal(t, "plugin", gt.PackageName())
		require.False(t, gt.IsPointer())
		require.True(t, gt.IsArray())
	})

	t.Run("pointer to array []*plugin.Table", func(t *testing.T) {
		gt := NewGoType("[]*plugin.Table")
		require.NotNil(t, gt)
		require.Equal(t, "[]*plugin.Table", gt.String())
		require.Equal(t, "Table", gt.TypeName())
		require.Equal(t, "plugin", gt.PackageName())
		require.True(t, gt.IsPointer())
		require.True(t, gt.IsArray())
	})
}
