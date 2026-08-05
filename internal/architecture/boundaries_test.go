package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// Allowed imports encode ADR-0001 boundaries for the bootstrap skeleton.
// Outer layers must not leak into domain/policy packages.
var forbiddenPrefixes = map[string][]string{
	"internal/policy": {
		"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver",
		"github.com/platformrelay/ibm-mq-mcp-server/internal/application",
		"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter",
	},
	"internal/mqadmin": {
		"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver",
		"github.com/platformrelay/ibm-mq-mcp-server/internal/application",
		"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter",
	},
	"internal/messaging": {
		"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver",
		"github.com/platformrelay/ibm-mq-mcp-server/internal/application",
		"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter",
	},
	"internal/adapter": {
		"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver",
	},
	"internal/application": {
		"github.com/platformrelay/ibm-mq-mcp-server/internal/mcpserver",
		"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter",
	},
	"internal/mcpserver": {
		"github.com/platformrelay/ibm-mq-mcp-server/internal/adapter",
	},
}

func TestPackageBoundaries(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	fset := token.NewFileSet()

	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(root, filepath.Dir(path))
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		forbidden, ok := forbiddenPrefixes[rel]
		if !ok {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("parse %s: %v", path, err)
			return nil
		}
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			for _, prefix := range forbidden {
				if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
					t.Errorf("%s imports forbidden package %s", path, importPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRequiredLayoutExists(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	required := []string{
		"cmd/ibm-mq-mcp",
		"internal/mcpserver",
		"internal/application",
		"internal/policy",
		"internal/mqadmin",
		"internal/messaging",
		"internal/adapter",
	}
	for _, dir := range required {
		path := filepath.Join(root, dir)
		entries, err := filepath.Glob(filepath.Join(path, "*.go"))
		if err != nil || len(entries) == 0 {
			t.Errorf("expected at least one .go file in %s", dir)
		}
	}
}
