//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
)

type walker struct {
	nodes []ast.Node
}

func (w *walker) Enter(node ast.Node) ast.Visitor {
	w.nodes = append(w.nodes, node)
	return w
}

func (w *walker) Exit(node ast.Node) {
}

func getAllNodes(node ast.Node) []ast.Node {
	w := &walker{}
	ast.Walk(w, node)
	return w.nodes
}

func getCalls(node ast.Node, name string) []*ast.CallExpression {
	nodes := []*ast.CallExpression{}
	for _, n := range getAllNodes(node) {
		if ce, ok := n.(*ast.CallExpression); ok {
			var calleeName string
			switch callee := ce.Callee.(type) {
			case *ast.Identifier:
				calleeName = callee.Name
			case *ast.DotExpression:
				calleeName = callee.Identifier.Name
			default:
				continue
			}
			if calleeName == name {
				nodes = append(nodes, ce)
			}
		}
	}
	return nodes
}

func getPropertyValue(node ast.Node, key string) ast.Expression {
	for _, p := range node.(*ast.ObjectLiteral).Value {
		if p.Key == key {
			return p.Value
		}
	}
	return nil
}

type operation struct {
	startLine   int
	startColumn int
	endLine     int
	endColumn   int
	text        []string
}

type check struct {
	before     []string
	operations []operation
	after      []string
}

type test struct {
	description string
	checks      []check
}

func stringSliceToGoSource(slice []string) string {
	var b strings.Builder
	b.WriteString("[]string{\n")
	for _, s := range slice {
		b.WriteString(fmt.Sprintf("%#v,\n", s))
	}
	b.WriteString("}")
	return b.String()
}

func testToGoTest(test test, name string) string {
	var b strings.Builder

	b.WriteString("func Test")
	b.WriteString(name)
	b.WriteString("(t *testing.T) {\n")

	for _, c := range test.checks {
		b.WriteString("check(\n")
		b.WriteString("t,\n")
		b.WriteString(fmt.Sprintf("%v,\n", stringSliceToGoSource(c.before)))
		b.WriteString("[]operation{\n")
		for _, op := range c.operations {
			b.WriteString("operation{\n")
			b.WriteString(fmt.Sprintf("start: Loc{%v, %v},\n", op.startColumn, op.startLine))
			b.WriteString(fmt.Sprintf("end: Loc{%v, %v},\n", op.endColumn, op.endLine))
			b.WriteString(fmt.Sprintf("text: %v,\n", stringSliceToGoSource(op.text)))
			b.WriteString("},\n")
		}
		b.WriteString("},\n")
		b.WriteString(fmt.Sprintf("%v,\n", stringSliceToGoSource(c.after)))
		b.WriteString(")\n")
	}

	b.WriteString("}\n")

	return b.String()
}

func nodeToStringSlice(node ast.Node) []string {
	var result []string
	for _, s := range node.(*ast.ArrayLiteral).Value {
		result = append(result, s.(*ast.StringLiteral).Value)
	}
	return result
}

func nodeToStringSlice2(node ast.Node) []string {
	var result []string
	for _, o := range node.(*ast.ArrayLiteral).Value {
		result = append(result, getPropertyValue(o, "text").(*ast.StringLiteral).Value)
	}
	return result
}

func nodeToInt(node ast.Node) int {
	return int(node.(*ast.NumberLiteral).Value.(int64))
}

func getChecks(node ast.Node) []check {
	checks := []check{}

	for _, ce := range getCalls(node, "testApplyEdits") {
		if len(ce.ArgumentList) != 3 {
			// Wrong function
			continue
		}

		before := nodeToStringSlice2(ce.ArgumentList[0])
		after := nodeToStringSlice2(ce.ArgumentList[2])

		var operations []operation
		for _, op := range ce.ArgumentList[1].(*ast.ArrayLiteral).Value {
			args := getPropertyValue(op, "range").(*ast.NewExpression).ArgumentList
			operations = append(operations, operation{
				startLine:   nodeToInt(args[0]) - 1,
				startColumn: nodeToInt(args[1]) - 1,
				endLine:     nodeToInt(args[2]) - 1,
				endColumn:   nodeToInt(args[3]) - 1,
				text:        []string{getPropertyValue(op, "text").(*ast.StringLiteral).Value},
			})
		}

		checks = append(checks, check{before, operations, after})
	}

	for _, ce := range getCalls(node, "testApplyEditsWithSyncedModels") {
		if len(ce.ArgumentList) > 3 && ce.ArgumentList[3].(*ast.BooleanLiteral).Value {
			// inputEditsAreInvalid == true
			continue
		}

		before := nodeToStringSlice(ce.ArgumentList[0])
		after := nodeToStringSlice(ce.ArgumentList[2])

		var operations []operation
		for _, op := range getCalls(ce.ArgumentList[1], "editOp") {
			operations = append(operations, operation{
				startLine:   nodeToInt(op.ArgumentList[0]) - 1,
				startColumn: nodeToInt(op.ArgumentList[1]) - 1,
				endLine:     nodeToInt(op.ArgumentList[2]) - 1,
				endColumn:   nodeToInt(op.ArgumentList[3]) - 1,
				text:        nodeToStringSlice(op.ArgumentList[4]),
			})
		}

		checks = append(checks, check{before, operations, after})
	}

	return checks
}

func getTests(node ast.Node) []test {
	tests := []test{}
	for _, ce := range getCalls(node, "test") {
		description := ce.ArgumentList[0].(*ast.StringLiteral).Value
		body := ce.ArgumentList[1].(*ast.FunctionLiteral).Body
		checks := getChecks(body)
		if len(checks) > 0 {
			tests = append(tests, test{description, checks})
		}
	}
	return tests
}

func main() {
	var tests []test

	for _, filename := range os.Args[1:] {
		source, err := os.ReadFile(filename)
		if err != nil {
			log.Fatalln(err)
		}

		program, err := parser.ParseFile(nil, "", source, parser.IgnoreRegExpErrors)
		if err != nil {
			log.Fatalln(err)
		}

		tests = append(tests, getTests(program)...)
	}

	if len(tests) == 0 {
		log.Fatalln("no tests found!")
	}

	fmt.Println("// This file is generated from VSCode model tests by the testgen tool.")
	fmt.Println("// DO NOT EDIT THIS FILE BY HAND; your changes will be overwritten!\n")
	fmt.Println("package buffer")
	fmt.Println(`import "testing"`)

	re := regexp.MustCompile(`[^\w]`)
	usedNames := map[string]bool{}

	for _, test := range tests {
		name := strings.Title(strings.ToLower(test.description))
		name = re.ReplaceAllLiteralString(name, "")
		if name == "" {
			name = "Unnamed"
		}
		if usedNames[name] {
			for i := 2; ; i++ {
				newName := fmt.Sprintf("%v_%v", name, i)
				if !usedNames[newName] {
					name = newName
					break
				}
			}
		}
		usedNames[name] = true

		fmt.Println(testToGoTest(test, name))
	}
}
