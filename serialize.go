/*
	This file is a serialization code generator
	Usage:
	//go:generate go run serialize.go -type=A,B,C
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	funcTemplateSerializeLen = iota
	funcTemplateSerialize
	funcTemplateDeserialize
)

type funcTemplate int

const (
	// fileHeader is a generated file header.
	// Arguments to format are
	//	[1]: package name
	//	[2]: source file name
	fileHeader = `// Code generated by serialize.
// source: %[2]s
// DO NOT EDIT!

package %[1]s

import (
	"encoding/binary"
	"unsafe"
)

var _ = unsafe.Sizeof(0)
`

	// funcSerialize is definition for Serialize function.
	// Arguments to format are
	//	[1]: type name
	funcSerializeLen = `func (i *%[1]s) SerializeLen() (n int) {
`
	funcSerializeLenEnd = `return
}
`

	// funcSerialize is definition for Serialize function.
	// Arguments to format are
	//	[1]: type name
	funcSerialize = `func (i *%[1]s) Serialize(b []byte, order binary.ByteOrder) {
`
	funcSerializeEnd = `}
`

	// funcDeserialize is definition for Deserialize function.
	// Arguments to format are
	//	[1]: type name
	funcDeserialize = `func (i *%[1]s) Deserialize(b []byte, order binary.ByteOrder) {
`
	funcDeserializeEnd = `}
`

	// blankLine is a blank line.
	blankLine = "\n"
)

// put expressions
// Arguments to format are
//	[1]: var name
const (
	// exprPut8 appends a 8-bit integer field in byte slice.
	exprPut8 = `b[0]=byte(%[1]s)
b=b[1:]
`

	// exprPut16 appends a 16-bit integer field in byte slice.
	exprPut16 = `order.PutUint16(b,uint16(%[1]s))
b=b[2:]
`

	// exprPut32 appends a 32-bit integer field in byte slice.
	exprPut32 = `order.PutUint32(b,uint32(%[1]s))
b=b[4:]
`

	// exprPut64 appends a 64-bit integer field in byte slice.
	exprPut64 = `order.PutUint64(b,uint64(%[1]s))
b=b[8:]
`

	// exprPutLen appends a 32-bit length field in byte slice.
	exprPutLen = `order.PutUint32(b,uint32(len(%[1]s)))
b=b[4:]
`

	// exprPutUnsafe appends a field as its in-memory-representation. This may be arch specific.
	exprPutUnsafe = `b=b[copy(b,(*(*[unsafe.Sizeof(%[1]s)]byte)(unsafe.Pointer(&%[1]s)))[:]):]
`
)

// load expressions
// Arguments to format are
//	[1]: var name
//	[2]: type name
const (
	// exprLoad8 loads a 8-bit integer field from byte slice.
	exprLoad8 = `%[1]s=%[2]s(b[0])
b=b[1:]
`

	// exprLoad16 loads a 16-bit integer field from byte slice.
	exprLoad16 = `%[1]s=%[2]s(order.Uint16(b[:2]))
b=b[2:]
`

	// exprLoad32 loads a 32-bit integer field from byte slice.
	exprLoad32 = `%[1]s=%[2]s(order.Uint32(b[:4]))
b=b[4:]
`

	// exprLoad64 loads a 64-bit integer field from byte slice.
	exprLoad64 = `%[1]s=%[2]s(order.Uint64(b[:8]))
b=b[8:]
`

	// exprLoadLen appends a 32-bit length field in byte slice.
	exprLoadLen = `{
l:=uint(order.Uint32(b[:4]))
b=b[4:]
`
	exprLoadLenEnd = "}\n"

	// exprLoadUnsafe loads a field as its in-memory-representation. This may be arch specific.
	exprLoadUnsafe = `%[1]s=*(*%[2]s)(unsafe.Pointer(&b[0]))
b=b[unsafe.Sizeof(%[1]s):]
`
)

type context struct {
	source  string
	fileSet *token.FileSet
	outputs map[string]*bytes.Buffer
	file    *ast.File
	pkg     *types.Package
}

func (t *context) initOutput(typeName string) *bytes.Buffer {
	name := fmt.Sprintf("%s_serialize.gen.go", typeName)
	if o, ok := t.outputs[name]; ok {
		return o
	}
	o := &bytes.Buffer{}
	o.WriteString(fmt.Sprintf(fileHeader, t.pkg.Name(), t.source))
	t.outputs[name] = o
	return o
}

// printNode converts a ast node to string
func (t *context) printNode(node interface{}) string {
	b := &strings.Builder{}
	printer.Fprint(b, t.fileSet, node)
	return b.String()
}

func (t *context) genField(out *bytes.Buffer, v string, expr ast.Expr, template funcTemplate) {
	switch typ := expr.(type) {
	case *ast.Ident:
		typeName := typ.String()
		switch typeName {
		case "bool":
		case "int8", "uint8", "byte":
			switch template {
			case funcTemplateSerializeLen:
				out.WriteString(`n++
`)
			case funcTemplateSerialize:
				out.WriteString(fmt.Sprintf(exprPut8, v))
			case funcTemplateDeserialize:
				out.WriteString(fmt.Sprintf(exprLoad8, v, typeName))
			}
		case "int16", "uint16":
			switch template {
			case funcTemplateSerializeLen:
				out.WriteString(`n+=2
`)
			case funcTemplateSerialize:
				out.WriteString(fmt.Sprintf(exprPut16, v))
			case funcTemplateDeserialize:
				out.WriteString(fmt.Sprintf(exprLoad16, v, typeName))
			}
		case "int32", "uint32", "rune":
			switch template {
			case funcTemplateSerializeLen:
				out.WriteString(`n+=4
`)
			case funcTemplateSerialize:
				out.WriteString(fmt.Sprintf(exprPut32, v))
			case funcTemplateDeserialize:
				out.WriteString(fmt.Sprintf(exprLoad32, v, typeName))
			}
		case "int64", "uint64", "int", "uint":
			switch template {
			case funcTemplateSerializeLen:
				out.WriteString(`n+=8
`)
			case funcTemplateSerialize:
				out.WriteString(fmt.Sprintf(exprPut64, v))
			case funcTemplateDeserialize:
				out.WriteString(fmt.Sprintf(exprLoad64, v, typeName))
			}
		case "float32", "float64", "complex64", "complex128":
			switch template {
			case funcTemplateSerializeLen:
				out.WriteString(fmt.Sprintf(`n+=int(unsafe.Sizeof(%[1]s))
`, v))
			case funcTemplateSerialize:
				out.WriteString(fmt.Sprintf(exprPutUnsafe, v))
			case funcTemplateDeserialize:
				out.WriteString(fmt.Sprintf(exprLoadUnsafe, v, typeName))
			}
		case "string":
			switch template {
			case funcTemplateSerializeLen:
				out.WriteString(fmt.Sprintf(`n+=4+len(%[1]s)
`, v))
			case funcTemplateSerialize:
				out.WriteString(fmt.Sprintf(`_=b[:4+len(%[1]s)]
order.PutUint32(b,uint32(len(%[1]s)))
b=b[4+copy(b[4:],%[1]s):]
`, v))
			case funcTemplateDeserialize:
				out.WriteString(fmt.Sprintf(`{
n:=order.Uint32(b)
%[1]s=string(b[4:4+n])
b=b[4+n:]
}
`, v))
			}
		default:
			switch template {
			case funcTemplateSerializeLen:
				out.WriteString(fmt.Sprintf(`n+=%[1]s.SerializeLen()
`, v))
			case funcTemplateSerialize:
				t.pkg.Scope().Lookup(typeName)
				out.WriteString(fmt.Sprintf(`%[1]s.Serialize(b,order)
b=b[%[1]s.SerializeLen():]
`, v))
			case funcTemplateDeserialize:
				out.WriteString(fmt.Sprintf(`%[1]s.Deserialize(b,order)
b=b[%[1]s.SerializeLen():]
`, v))
			}
		}
	case *ast.ArrayType:
		switch template {
		case funcTemplateSerializeLen:
			if typ.Len != nil { // array
				out.WriteString(fmt.Sprintf(`for _,element:=range %[1]s {
`, v))
				t.genField(out, "element", typ.Elt, template)
				out.WriteString("}\n")
			} else { // slice
				out.WriteString(`n+=4
`)
				out.WriteString(fmt.Sprintf(`for _,element:=range %[1]s {
`, v))
				t.genField(out, "element", typ.Elt, template)
				out.WriteString("}\n")
			}
		case funcTemplateSerialize:
			if typ.Len != nil { // array
				out.WriteString(fmt.Sprintf(`for _,element:=range %[1]s {
`, v))
				t.genField(out, "element", typ.Elt, template)
				out.WriteString("}\n")
			} else { // slice
				out.WriteString(fmt.Sprintf(exprPutLen, v))
				out.WriteString(fmt.Sprintf(`for _,element:=range %[1]s {
`, v))
				t.genField(out, "element", typ.Elt, template)
				out.WriteString("}\n")
			}
		case funcTemplateDeserialize:
			if typ.Len != nil { // array
				out.WriteString(fmt.Sprintf(`for k:=uint(0);k<%[1]s;k++{
`, t.printNode(typ.Len)))
				t.genField(out, fmt.Sprintf("%[1]s[k]", v), typ.Elt, template)
				out.WriteString("}\n")
			} else { // slice
				out.WriteString(exprLoadLen)
				out.WriteString(fmt.Sprintf(`%[1]s=make(%[2]s,l,l)
b=b[l:]
for k:=uint(0);k<l;k++{
`, v, t.printNode(typ)))
				t.genField(out, fmt.Sprintf("%[1]s[k]", v), typ.Elt, template)
				out.WriteString("}\n")
				out.WriteString(exprLoadLenEnd)
			}
		}
	case *ast.MapType:
		switch template {
		case funcTemplateSerializeLen:
			out.WriteString(`n+=4
`)
			out.WriteString(fmt.Sprintf(`for key,value:=range %[1]s {
`, v))
			t.genField(out, "key", typ.Key, template)
			t.genField(out, "value", typ.Value, template)
			out.WriteString("}\n")
		case funcTemplateSerialize:
			out.WriteString(fmt.Sprintf(exprPutLen, v))
			out.WriteString(fmt.Sprintf(`for key,value:=range %[1]s {
`, v))
			t.genField(out, "key", typ.Key, template)
			t.genField(out, "value", typ.Value, template)
			out.WriteString("}\n")
		case funcTemplateDeserialize:
			out.WriteString(exprLoadLen)
			out.WriteString(fmt.Sprintf(`m:=make(map[%[1]s]%[2]s)
for k:=uint(0);k<l;k++{
var key %[1]s
var value %[2]s
`, t.printNode(typ.Key), t.printNode(typ.Value)))
			t.genField(out, "key", typ.Key, template)
			t.genField(out, "value", typ.Value, template)
			out.WriteString(fmt.Sprintf(`m[key]=value
}
%[1]s=m
`, v))
			out.WriteString(exprLoadLenEnd)
		}
	case *ast.StructType:
		t.genStruct(out, v, typ, template)
	default:
		log.Fatalln("unsupported type:", t.printNode(expr))
	}
}

func (t *context) genStruct(out *bytes.Buffer, v string, s *ast.StructType, template funcTemplate) {
	for _, f := range s.Fields.List {
		if len(f.Names) > 0 {
			for _, n := range f.Names {
				t.genField(out, fmt.Sprintf("%s.%s", v, n.String()), f.Type, template)
			}
		} else {
			t.genField(out, fmt.Sprintf("%s.%s", v, t.printNode(f.Type)), f.Type, template)
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
	out.WriteString(fmt.Sprintf(funcSerializeLen, typeName))
	if structType, ok := typeSpec.Type.(*ast.StructType); ok {
		t.genStruct(out, "i", structType, funcTemplateSerializeLen)
	} else {
		t.genField(out, "*i", typeSpec.Type, funcTemplateSerializeLen)
	}
	out.WriteString(funcSerializeLenEnd)

	out.WriteString(blankLine)
	out.WriteString(fmt.Sprintf(funcSerialize, typeName))
	if structType, ok := typeSpec.Type.(*ast.StructType); ok {
		t.genStruct(out, "i", structType, funcTemplateSerialize)
	} else {
		t.genField(out, "*i", typeSpec.Type, funcTemplateSerialize)
	}
	out.WriteString(funcSerializeEnd)

	out.WriteString(blankLine)
	out.WriteString(fmt.Sprintf(funcDeserialize, typeName))
	if structType, ok := typeSpec.Type.(*ast.StructType); ok {
		t.genStruct(out, "i", structType, funcTemplateDeserialize)
	} else {
		t.genField(out, "f:", typeSpec.Type, funcTemplateDeserialize)
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
		source:  fileName,
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
