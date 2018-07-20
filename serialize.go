/*
	This file is a serialization code generator
	Usage:
	//go:generate go run serialize.go -type=A,B,C
*/
package main

import (
	"flag"
	"log"
	"os"
	"go/token"
	"bytes"
	"strings"
	"go/parser"
	"go/ast"
	"go/types"
	"go/importer"
	"go/format"
	"io/ioutil"
	"path/filepath"
	"fmt"
	"go/printer"
)

const (
	// fileHeader is a generated file header.
	// Arguments to format are
	//	[1]: package name
	fileHeader = `package %[1]s

import "unsafe"

var _ = unsafe.Sizeof(0)
`

	// funcSerialize is definition for Serialize function.
	// Arguments to format are
	//	[1]: type name
	funcSerialize = `func (i *%[1]s) Serialize(b []byte) []byte {
`
	funcSerializeEnd = `return b
}
`

	// funcDeserialize is definition for Deserialize function.
	// Arguments to format are
	//	[1]: type name
	funcDeserialize = `func (i *%[1]s) Deserialize(b []byte) []byte {
`
	funcDeserializeEnd = `return b
}
`

	// blankLine is a blank line.
	blankLine = "\n"
)

// append expressions
// Arguments to format are
//	[1]: type name
const (
	// exprAppend8 appends a 8-bit integer field in byte slice.
	exprAppend8 = `b=append(b,byte(%[1]s))
`

	// exprAppend16 appends a 16-bit integer field in byte slice.
	exprAppend16 = `b=append(b,byte(%[1]s>>8),byte(%[1]s))
`

	// exprAppend32 appends a 32-bit integer field in byte slice.
	exprAppend32 = `b=append(b,byte(%[1]s>>24),byte(%[1]s>>16),byte(%[1]s>>8),byte(%[1]s))
`

	// exprAppend64 appends a 64-bit integer field in byte slice.
	exprAppend64 = `b=append(b,byte(%[1]s>>56),byte(%[1]s>>48),byte(%[1]s>>40),byte(%[1]s>>32),byte(%[1]s>>24),byte(%[1]s>>16),byte(%[1]s>>8),byte(%[1]s))
`

	// exprAppendLen appends a 32-bit length field in byte slice.
	exprAppendLen = `{
l:=len(%[1]s)
b=append(b,byte(l>>24),byte(l>>16),byte(l>>8),byte(l))
`
	exprAppendLenEnd = "}\n"

	// exprAppendUnsafe appends a field as its in-memory-representation. This may be arch specific.
	exprAppendUnsafe = `b=append(b,(*(*[unsafe.Sizeof(%[1]s)]byte)(unsafe.Pointer(&%[1]s)))[:]...)
`
)

// load expressions
// Arguments to format are
//	[1]: var name
//	[1]: type name
const (
	// exprLoad8 loads a 8-bit integer field from byte slice.
	exprLoad8 = `%[1]s=%[2]s(b[0])
b=b[1:]
`

	// exprLoad16 loads a 16-bit integer field from byte slice.
	exprLoad16 = `%[1]s=%[2]s(b[1])|%[2]s(b[0])<<8
b=b[2:]
`

	// exprLoad32 loads a 32-bit integer field from byte slice.
	exprLoad32 = `%[1]s=%[2]s(b[3])|%[2]s(b[2])<<8|%[2]s(b[1])<<16|%[2]s(b[0])<<24
b=b[4:]
`

	// exprLoad64 loads a 64-bit integer field from byte slice.
	exprLoad64 = `%[1]s=%[2]s(b[7])|%[2]s(b[6])<<8|%[2]s(b[5])<<16|%[2]s(b[4])<<24|%[2]s(b[3])<<32|%[2]s(b[2])<<40|%[2]s(b[1])<<48|%[2]s(b[0])<<56
b=b[8:]
`

	// exprLoadLen appends a 32-bit length field in byte slice.
	exprLoadLen = `{
l:=uint(b[3])|uint(b[2])<<8|uint(b[1])<<16|uint(b[0])<<24
b=b[4:]
`
	exprLoadLenEnd = "}\n"

	// exprLoadUnsafe loads a field as its in-memory-representation. This may be arch specific.
	exprLoadUnsafe = `%[1]s=*(*%[2]s)(unsafe.Pointer(&b[0]))
b=b[unsafe.Sizeof(%[1]s):]
`
)

type context struct {
	dir     string
	fileSet *token.FileSet
	outputs map[string]*bytes.Buffer
	file    *ast.File
	pkg     *types.Package
}

func (t *context) initOutput(typeName string) *bytes.Buffer {
	name := filepath.Join(filepath.Dir(t.file.Name.Name), fmt.Sprintf("%s_serialize.gen.go", typeName))
	if o, ok := t.outputs[name]; ok {
		return o
	}
	o := &bytes.Buffer{}
	o.WriteString(fmt.Sprintf(fileHeader, t.pkg.Name()))
	t.outputs[name] = o
	return o
}

// printNode converts a ast node to string
func (t *context) printNode(node interface{}) string {
	b := &strings.Builder{}
	printer.Fprint(b, t.fileSet, node)
	return b.String()
}

func (t *context) genField(out *bytes.Buffer, v string, expr ast.Expr, serializing bool) {
	switch typ := expr.(type) {
	case *ast.Ident:
		typeName := typ.String()
		switch typeName {
		case "bool":
		case "int8", "uint8", "byte":
			if serializing {
				out.WriteString(fmt.Sprintf(exprAppend8, v))
			} else {
				out.WriteString(fmt.Sprintf(exprLoad8, v, typeName))
			}
		case "int16", "uint16":
			if serializing {
				out.WriteString(fmt.Sprintf(exprAppend16, v))
			} else {
				out.WriteString(fmt.Sprintf(exprLoad16, v, typeName))
			}
		case "int32", "uint32", "rune":
			if serializing {
				out.WriteString(fmt.Sprintf(exprAppend32, v))
			} else {
				out.WriteString(fmt.Sprintf(exprLoad32, v, typeName))
			}
		case "int64", "uint64", "int", "uint":
			if serializing {
				out.WriteString(fmt.Sprintf(exprAppend64, v))
			} else {
				out.WriteString(fmt.Sprintf(exprLoad64, v, typeName))
			}
		case "float32", "float64", "complex64", "complex128":
			if serializing {
				out.WriteString(fmt.Sprintf(exprAppendUnsafe, v))
			} else {
				out.WriteString(fmt.Sprintf(exprLoadUnsafe, v, typeName))
			}
		case "string":
			if serializing {
				out.WriteString(fmt.Sprintf(exprAppendLen, v))
				out.WriteString(fmt.Sprintf(`b=append(b,[]byte(%[1]s)...)
`, v))
				out.WriteString(exprAppendLenEnd)
			} else {
				out.WriteString(exprLoadLen)
				out.WriteString(fmt.Sprintf(`%[1]s=string(b[:l])
b=b[l:]
`, v))
				out.WriteString(exprLoadLenEnd)
			}
		default:
			if serializing {
				t.pkg.Scope().Lookup(typeName)
				out.WriteString(fmt.Sprintf(`b=%[1]s.Serialize(b)
`, v))
			} else {
				out.WriteString(fmt.Sprintf(`b=%[1]s.Deserialize(b)
`, v))
			}
		}
	case *ast.ArrayType:
		if serializing {
			out.WriteString(fmt.Sprintf(exprAppendLen, v))
			out.WriteString(fmt.Sprintf(`for _,element:=range %[1]s {
`, v))
			t.genField(out, "element", typ.Elt, serializing)
			out.WriteString("}\n")
			out.WriteString(exprAppendLenEnd)
		} else {
			out.WriteString(exprLoadLen)
			out.WriteString(fmt.Sprintf(`%[1]s=make(%[2]s,l,l)
b=b[l:]
`, v, t.printNode(typ)))
			out.WriteString(exprLoadLenEnd)
		}
	case *ast.MapType:
		if serializing {
			out.WriteString(fmt.Sprintf(exprAppendLen, v))
			out.WriteString(fmt.Sprintf(`for key,value:=range %[1]s {
`, v))
			t.genField(out, "key", typ.Key, serializing)
			t.genField(out, "value", typ.Value, serializing)
			out.WriteString("}\n")
			out.WriteString(exprAppendLenEnd)
		} else {
			out.WriteString(exprLoadLen)
			out.WriteString(fmt.Sprintf(`m:=make(map[%[1]s]%[2]s)
for k:=uint(0);k<l;k++{
var key %[1]s
var value %[2]s
`, t.printNode(typ.Key), t.printNode(typ.Value)))
			t.genField(out, "key", typ.Key, serializing)
			t.genField(out, "value", typ.Value, serializing)
			out.WriteString(fmt.Sprintf(`m[key]=value
}
%[1]s=m
`, v))
			out.WriteString(exprLoadLenEnd)
		}
	case *ast.StructType:
		t.genStruct(out, v, typ, serializing)
	default:
		log.Fatalln("unsupported type:", t.printNode(expr))
	}
}

func (t *context) genStruct(out *bytes.Buffer, v string, s *ast.StructType, serializing bool) {
	for _, f := range s.Fields.List {
		if len(f.Names) > 0 {
			for _, n := range f.Names {
				t.genField(out, fmt.Sprintf("%s.%s", v, n.String()), f.Type, serializing)
			}
		} else {
			t.genField(out, fmt.Sprintf("%s.%s", v, t.printNode(f.Type)), f.Type, serializing)
		}
	}
}

func (t *context) generate(typeName string) {
	obj := t.file.Scope.Lookup(typeName)
	if obj == nil {
		log.Fatalln("type not found:", typeName)
	}
	if obj.Kind != ast.Typ {
		log.Fatalln(typeName, "is not a type")
	}
	typeSpec := obj.Decl.(*ast.TypeSpec)

	log.Printf("generating for %s.%s", t.pkg.Name(), typeName)

	out := t.initOutput(typeName)
	out.WriteString(blankLine)
	out.WriteString(fmt.Sprintf(funcSerialize, typeName))
	if structType, ok := typeSpec.Type.(*ast.StructType); ok {
		t.genStruct(out, "i", structType, true)
	} else {
		t.genField(out, "*i", typeSpec.Type, true)
	}
	out.WriteString(funcSerializeEnd)

	out.WriteString(blankLine)
	out.WriteString(fmt.Sprintf(funcDeserialize, typeName))
	if structType, ok := typeSpec.Type.(*ast.StructType); ok {
		t.genStruct(out, "i", structType, false)
	} else {
		t.genField(out, "f:", typeSpec.Type, false)
		out.WriteString(fmt.Sprintf("*i=%[1]s(f)\n", typeName))
	}
	out.WriteString(funcDeserializeEnd)
}

var (
	typeFlag = flag.String("type", "", "type names to generate")
)

func main() {
	flag.Parse()
	log.SetPrefix("[serialize_gen] ")
	typeNames := strings.Split(*typeFlag, ",")
	if len(typeNames) < 1 {
		log.Fatalln("incorrect -type")
	}

	pkgName := os.Getenv("GOPACKAGE")
	fileName := os.Getenv("GOFILE")

	log.Println("package", pkgName)

	t := context{
		fileSet: token.NewFileSet(),
		outputs: make(map[string]*bytes.Buffer),
		dir:     filepath.Dir(fileName),
	}

	pkgs, err := parser.ParseDir(t.fileSet, ".", nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("error parsing %s: %v", fileName, err)
	}

	files := make([]*ast.File, 0)
	for _, pkg := range pkgs {
		if pkg.Name == pkgName {
			for name, file := range pkg.Files {
				files = append(files, file)
				if name == fileName {
					t.file = file
				}
			}
			break
		}
	}

	config := types.Config{Importer: importer.Default(), FakeImportC: true}
	t.pkg, err = config.Check("", t.fileSet, files, &types.Info{Defs: make(map[*ast.Ident]types.Object)})
	if err != nil {
		log.Fatalf("type check error: %v", err)
	}

	for _, typeName := range typeNames {
		t.generate(typeName)
	}

	for file, buf := range t.outputs {
		b, err := format.Source(buf.Bytes())
		if err != nil {
			log.Printf("error formatting %s: %v", file, err)
			b = buf.Bytes()
		}
		err = ioutil.WriteFile(file, b, 0644)
		if err != nil {
			log.Fatalf("error writting %s: %v", file, err)
		}
	}
	log.Println("OK")
}