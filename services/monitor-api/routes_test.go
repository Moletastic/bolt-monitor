package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMonitorRoutesMatchHandleRequestDispatch(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate route coverage test")
	}
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filepath.Join(filepath.Dir(testFile), "handler.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse handler dispatch: %v", err)
	}

	dispatchHandlers := make(map[string]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		function, ok := node.(*ast.FuncDecl)
		if !ok || function.Name.Name != "handleRequest" {
			return true
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			switchStatement, ok := node.(*ast.SwitchStmt)
			if !ok {
				return true
			}
			ast.Inspect(switchStatement, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				receiver, ok := selector.X.(*ast.Ident)
				if ok && receiver.Name == "h" {
					dispatchHandlers[selector.Sel.Name] = struct{}{}
				}
				return true
			})
			return false
		})
		return false
	})

	inventoryHandlers := make(map[string]struct{})
	for _, route := range monitorRoutes {
		if route.Handler == "" {
			t.Errorf("inventory route %s %s has no dispatch handler", route.Method, route.Path)
			continue
		}
		inventoryHandlers[route.Handler] = struct{}{}
		if _, ok := dispatchHandlers[route.Handler]; !ok {
			t.Errorf("inventory route %s %s expects h.%s in handleRequest", route.Method, route.Path, route.Handler)
		}
	}
	for handler := range dispatchHandlers {
		if _, ok := inventoryHandlers[handler]; !ok {
			t.Errorf("handleRequest dispatch h.%s is absent from monitorRoutes inventory", handler)
		}
	}
}
