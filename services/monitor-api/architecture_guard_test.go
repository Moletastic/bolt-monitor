package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProductionCompositionRootsOwnEnvironmentAndAWSClientConstruction(t *testing.T) {
	root := repositoryRoot(t)
	allowed := map[string]struct{}{
		"services/check-runtime/main.go":                        {},
		"services/escalation-runtime/main.go":                   {},
		"services/monitor-api/main.go":                          {},
		"services/monitor-api/inline_channel_migration_main.go": {},
		"services/monitor-api/cmd/audit-index-backfill/main.go": {},
	}

	forEachProductionServiceFile(t, root, func(path string, file *ast.File, fset *token.FileSet) {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		_, isCompositionRoot := allowed[filepath.ToSlash(rel)]
		aliases := importAliases(file)

		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || isCompositionRoot {
				return true
			}
			if isEnvironmentRead(call, aliases) || isAWSClientConstruction(call, aliases) {
				position := fset.Position(call.Pos())
				t.Errorf("%s:%d: environment reads and concrete AWS client construction belong in a composition root", rel, position.Line)
			}
			return true
		})
	})
}

func TestDeclaredQueriesDoNotWrite(t *testing.T) {
	root := repositoryRoot(t)
	forEachProductionServiceFile(t, root, func(path string, file *ast.File, fset *token.FileSet) {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || !isQueryOperation(function.Name.Name) {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || !isWriteCall(call) {
					return true
				}
				position := fset.Position(call.Pos())
				t.Errorf("%s:%d: query %s must not write or publish side effects", rel, position.Line, function.Name.Name)
				return true
			})
		}
	})
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, path, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture guard path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(path), "../.."))
}

func forEachProductionServiceFile(t *testing.T, root string, visit func(string, *ast.File, *token.FileSet)) {
	t.Helper()
	err := filepath.WalkDir(filepath.Join(root, "services"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		visit(path, file, fset)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func importAliases(file *ast.File) map[string]string {
	aliases := make(map[string]string, len(file.Imports))
	for _, imported := range file.Imports {
		path := strings.Trim(imported.Path.Value, `"`)
		name := filepath.Base(path)
		if imported.Name != nil {
			name = imported.Name.Name
		}
		aliases[name] = path
	}
	return aliases
}

func isEnvironmentRead(call *ast.CallExpr, aliases map[string]string) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || aliases[selectorPackage(selector)] != "os" {
		return false
	}
	return selector.Sel.Name == "Getenv" || selector.Sel.Name == "LookupEnv" || selector.Sel.Name == "Environ"
}

func isAWSClientConstruction(call *ast.CallExpr, aliases map[string]string) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	packagePath := aliases[selectorPackage(selector)]
	if packagePath == "bolt-monitor/shared/aws" {
		return strings.HasPrefix(selector.Sel.Name, "New") && strings.HasSuffix(selector.Sel.Name, "API")
	}
	return strings.HasPrefix(packagePath, "github.com/aws/aws-sdk-go-v2") &&
		(selector.Sel.Name == "LoadDefaultConfig" || strings.HasPrefix(selector.Sel.Name, "New"))
}

func selectorPackage(selector *ast.SelectorExpr) string {
	identifier, _ := selector.X.(*ast.Ident)
	if identifier == nil {
		return ""
	}
	return identifier.Name
}

func isQueryOperation(name string) bool {
	return strings.HasPrefix(name, "Get") || strings.HasPrefix(name, "List") || strings.HasPrefix(name, "Search")
}

func isWriteCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch selector.Sel.Name {
	case "PutItem", "UpdateItem", "DeleteItem", "TransactWriteItems", "SendMessage", "SendMessageBatch", "Publish", "CreateSchedule", "UpdateSchedule", "DeleteSchedule", "MigrateRouteInlineChannels", "writeTransaction", "replaceSearchIndex", "deleteSearchIndex":
		return true
	default:
		return false
	}
}
