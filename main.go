package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	mainFile *string = flag.String("main", "main.go", "use -main <mainfile>")
)

func main() {
	flag.Parse()
	dir , _ := filepath.Abs("./")
	parse := NewParser()
	parse.ParseApi(dir, *mainFile)
	swag := parse.swagger
	b, _ := json.MarshalIndent(swag, "", "")
	docs, _ := os.Create(path.Join(dir, "swagger.json"))
	defer docs.Close()
	docs.Write(b)
}

// Parser implements a parser for Go source files.
type Parser struct {
	// swagger represents the root document object for the API specification
	swagger *spec.Swagger

	//files is a map that stores map[real_go_file_path][astFile]
	files map[string]*ast.File

	// TypeDefinitions is a map that stores [package name][type name][*ast.TypeSpec]
	TypeDefinitions map[string]map[string]*ast.TypeSpec

	//registerTypes is a map that stores [refTypeName][*ast.TypeSpec]
	registerTypes map[string]*ast.TypeSpec

	//定义的类
	Definitions map[string]*ast.TypeSpec
}

type Operation struct {
	HttpMethod string
	Path       string
	spec.Operation

	parser *Parser // TODO: we don't need it
}

// NewOperation creates a new Operation with default properties.
// map[int]Response
func NewOperation() *Operation {
	return &Operation{
		HttpMethod: "get",
		Operation: spec.Operation{
			OperationProps: spec.OperationProps{},
		},
	}
}

func NewParser() *Parser {
	parser := &Parser{
		swagger: &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Info: &spec.Info{
					InfoProps: spec.InfoProps{
						Contact: &spec.ContactInfo{},
						License: &spec.License{},
					},
				},
				Paths: &spec.Paths{
					Paths: make(map[string]spec.PathItem),
				},
				Definitions: make(map[string]spec.Schema),
			},
		},
		files:           make(map[string]*ast.File),
		TypeDefinitions: make(map[string]map[string]*ast.TypeSpec),
		registerTypes:   make(map[string]*ast.TypeSpec),
		Definitions:     make(map[string]*ast.TypeSpec),
	}
	return parser
}

func (p *Parser) ParseApi(dir string, main string) {
	p.getAllGoFileInfo(dir)
	p.getApiInfo(filepath.Join(dir, main))

	for _, astFile := range p.files {
		p.ParseType(astFile)
	}
	for _, astFile := range p.files {
		p.ParseRouterApiInfo(astFile)
	}
	p.ParseDefinitions()
}

func (p *Parser) getAllGoFileInfo(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if ext := filepath.Ext(path); ext == ".go" && !strings.Contains(path, "vendor") {
			astFile, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ParseComments)
			if err != nil {
				panic(err)
			}
			p.files[path] = astFile
		}
		return nil
	})
}

func (p *Parser) getApiInfo(main string) {
	fileSet := token.NewFileSet()
	fileTree, err := parser.ParseFile(fileSet, main, nil, parser.ParseComments)
	if err != nil {
		log.Panicf("ParseGeneralApiInfo occur error:%+v", err)
	}
	p.swagger.Swagger = "2.0"
	if fileTree.Comments != nil {
		for _, comment := range fileTree.Comments {
			for _, commentLine := range strings.Split(comment.Text(), "\n") {
				attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
				switch attribute {
				case "@version":
					p.swagger.Info.Version = strings.TrimSpace(commentLine[len(attribute):])
				case "@title":
					p.swagger.Info.Title = strings.TrimSpace(commentLine[len(attribute):])
				case "@description":
					p.swagger.Info.Description = strings.TrimSpace(commentLine[len(attribute):])
				case "@termsofservice":
					p.swagger.Info.TermsOfService = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.name":
					p.swagger.Info.Contact.Name = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.email":
					p.swagger.Info.Contact.Email = strings.TrimSpace(commentLine[len(attribute):])
				case "@contact.url":
					p.swagger.Info.Contact.URL = strings.TrimSpace(commentLine[len(attribute):])
				case "@license.name":
					p.swagger.Info.License.Name = strings.TrimSpace(commentLine[len(attribute):])
				case "@license.url":
					p.swagger.Info.License.URL = strings.TrimSpace(commentLine[len(attribute):])
				case "@host":
					p.swagger.Host = strings.TrimSpace(commentLine[len(attribute):])
				case "@basepath":
					p.swagger.BasePath = strings.TrimSpace(commentLine[len(attribute):])
				case "@schemes":
					p.swagger.Schemes = GetSchemes(commentLine)
				case "@tags":
					p.swagger.Tags = append(p.swagger.Tags, GetTags(commentLine))
				}
			}
		}
	}
}

func (p *Parser) ParseType(file *ast.File) {
	if _, ok := p.TypeDefinitions[file.Name.String()]; !ok {
		p.TypeDefinitions[file.Name.String()] = make(map[string]*ast.TypeSpec)
	}
	for _, astDeclaration := range file.Decls {
		if generalDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && generalDeclaration.Tok == token.TYPE {
			for _, astSpec := range generalDeclaration.Specs {
				if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
					p.TypeDefinitions[file.Name.String()][typeSpec.Name.String()] = typeSpec
				}
			}
		}
	}
	/*
	*  在这里查找 def
	 */
	for _, astDeclaration := range file.Decls {
		switch astDec := astDeclaration.(type) {
		case *ast.GenDecl:
			if astDec.Doc != nil && astDec.Doc.List != nil {
				for _, comment := range astDec.Doc.List {
					text := strings.Trim(comment.Text, "/* ")
					if strings.Index(text, "@def") == 0 {
						genDecl := astDeclaration.(*ast.GenDecl)
						if genDecl.Tok == token.TYPE {
							if realType, ok := astDec.Specs[0].(*ast.TypeSpec); ok {
								p.Definitions[strings.TrimSpace(text[len("@def"):])] = realType
							}
						}
					}
				}
			}
		}
	}
}

func (p *Parser) ParseRouterApiInfo(file *ast.File) {
	for _, astDescription := range file.Decls {
		switch astDeclaration := astDescription.(type) {
		case *ast.FuncDecl:
			if astDeclaration.Doc != nil && astDeclaration.Doc.List != nil {
				operation := NewOperation() //for per 'function' comment, create a new 'Operation' object
				operation.parser = p
				for _, comment := range astDeclaration.Doc.List {
					if err := operation.ParseComment(comment.Text); err != nil {
						log.Panicf("ParseComment panic:%+v", err)
					}
				}
				var pathItem spec.PathItem
				var ok bool

				if pathItem, ok = p.swagger.Paths.Paths[operation.Path]; !ok {
					pathItem = spec.PathItem{}
				}
				switch strings.ToUpper(operation.HttpMethod) {
				case http.MethodGet:
					pathItem.Get = &operation.Operation
				case http.MethodPost:
					pathItem.Post = &operation.Operation
				case http.MethodDelete:
					pathItem.Delete = &operation.Operation
				case http.MethodPut:
					pathItem.Put = &operation.Operation
				case http.MethodPatch:
					pathItem.Patch = &operation.Operation
				case http.MethodHead:
					pathItem.Head = &operation.Operation
				case http.MethodOptions:
					pathItem.Options = &operation.Operation
				}

				p.swagger.Paths.Paths[operation.Path] = pathItem
			}
		}
	}
}

func (p *Parser) ParseDefinitions() {
	for refTypeName, typeSpec := range p.Definitions {
		var properties map[string]spec.Schema
		properties = make(map[string]spec.Schema)

		switch typeSpec.Type.(type) {
		case *ast.StructType:
			structDecl := typeSpec.Type.(*ast.StructType)
			fields := structDecl.Fields.List
			for _, field := range fields {
				if len(field.Names) > 0 {
					name := field.Names[0].Name
					propName := getPropertyName(field)
					name = snakeString(name)
					r := spec.Schema{
						SchemaProps: spec.SchemaProps{Type: []string{propName}},

					}
					if propName == "array" || propName == "object" {
						re := regexp.MustCompile(`swag\:\"(\w+)\"`)
						s := re.FindStringSubmatch(field.Tag.Value)
						if len(s) == 2 {
							r.Items = &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/" + s[1])},
									},
								},
							}
						}
					}
					properties[name] = r
				}
			}
		}

		p.swagger.Definitions[refTypeName] = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:       []string{"object"},
				Properties: properties,
			},
		}

	}
}

// GetSchemes parses swagger schemes for gived commentLine
func GetSchemes(commentLine string) []string {
	attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
	return strings.Split(strings.TrimSpace(commentLine[len(attribute):]), " ")
}

func GetTags(commentLine string) spec.Tag {
	attr := strings.Split(commentLine, " ")
	var tag = *&spec.Tag{}
	if len(attr) > 1 {
		tag.TagProps = spec.TagProps{
			Name: attr[1],
		}
	}
	if len(attr) > 2 {
		tag.TagProps.Description = attr[2]
	}
	return tag
}

// ParseComment parses comment for gived comment string and returns error if error occurs.
func (operation *Operation) ParseComment(comment string) error {
	commentLine := strings.TrimSpace(strings.TrimLeft(comment, "//"))
	if len(commentLine) == 0 {
		return nil
	}
	attribute := strings.Fields(commentLine)[0]
	switch strings.ToLower(attribute) {
	case "@description":
		operation.Description = strings.TrimSpace(commentLine[len(attribute):])
	case "@summary":
		operation.Summary = strings.TrimSpace(commentLine[len(attribute):])
	case "@id":
		operation.ID = strings.TrimSpace(commentLine[len(attribute):])
	case "@tag":
		operation.Tags = operation.ParseTagComment(commentLine)
	case "@accept":
		if err := operation.ParseAcceptComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@produce":
		if err := operation.ParseProduceComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@param":
		if err := operation.ParseParamComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	case "@success", "@failure":
		if err := operation.ParseResponseComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {

			if errWhenEmpty := operation.ParseEmptyResponseComment(strings.TrimSpace(commentLine[len(attribute):])); errWhenEmpty != nil {
				var errs []string
				errs = append(errs, err.Error())
				errs = append(errs, errWhenEmpty.Error())
				return fmt.Errorf(strings.Join(errs, "\n"))
			}
		}

	case "@router":
		if err := operation.ParseRouterComment(strings.TrimSpace(commentLine[len(attribute):])); err != nil {
			return err
		}
	}

	return nil
}

// Parse params return []string of param properties
// @Param	queryText		form	      string	  true		        "The email for login"
// 			[param name]    [paramType] [data type]  [is mandatory?]   [Comment]
// @Param   some_id     path    int     true        "Some ID"
// @param page query string false  "页码"
// @param name body model.ArticleTag true  "标签名称"
// @param name body
func (operation *Operation) ParseParamComment(commentLine string) error {
	paramString := commentLine

	re := regexp.MustCompile(`([-\w]+)[\s]+([\w]+)[\s]+([\S.]+)[\s]+([\w]+)[\s]+"([^"]+)"`)

	if matches := re.FindStringSubmatch(paramString); len(matches) != 6 {
		return fmt.Errorf("Can not parse param comment \"%s\", skipped.", paramString)
	} else {
		name := matches[1]
		paramType := matches[2]
		schemaType := matches[3]

		requiredText := strings.ToLower(matches[4])
		required := (requiredText == "true" || requiredText == "required")
		description := matches[5]

		var param spec.Parameter

		//five possible parameter types.
		switch paramType {
		case "query", "path":
			param = createParameter(paramType, description, name, schemaType, required)
		case "body", "formData":
			var schema string
			if paramType == "body" {
				schema = "object"
			} else {
				schema = "file"
			}
			param = createParameter(paramType, description, name, schema, required) // TODO: if Parameter types can be objects, but also primitives and arrays

			// TODO: this snippets have to extract out
			if strings.Index(strings.TrimSpace(schemaType), "@") == 0 {
				schemaType = schemaType[1:]
				if ref, ok := operation.parser.Definitions[schemaType]; ok {
					operation.parser.registerTypes[schemaType] = ref
				}
				param.Schema.Ref = spec.Ref{
					Ref: jsonreference.MustCreateRef("#definitions/" + schemaType),
				}
			}
		case "header": // TODO: support Header and Form
			param = createFormDataParameter(paramType, description, name, schemaType, required)
		}
		operation.Operation.Parameters = append(operation.Operation.Parameters, param)
	}

	return nil
}
func (operation *Operation) ParseAcceptComment(commentLine string) error {
	accepts := strings.Split(commentLine, ",")
	for _, a := range accepts {
		switch a {
		case "json", "application/json":
			operation.Consumes = append(operation.Consumes, "application/json")
		case "xml", "text/xml":
			operation.Consumes = append(operation.Consumes, "text/xml")
		case "plain", "text/plain":
			operation.Consumes = append(operation.Consumes, "text/plain")
		case "html", "text/html":
			operation.Consumes = append(operation.Consumes, "text/html")
		case "mpfd", "multipart/form-data":
			operation.Consumes = append(operation.Consumes, "multipart/form-data")
		default:
			operation.Consumes = append(operation.Consumes, "*/*")
		}
	}
	return nil
}

func (operation *Operation) ParseProduceComment(commentLine string) error {
	produces := strings.Split(commentLine, ",")
	for _, a := range produces {
		switch a {
		case "json", "application/json":
			operation.Produces = append(operation.Produces, "application/json")
		case "xml", "text/xml":
			operation.Produces = append(operation.Produces, "text/xml")
		case "plain", "text/plain":
			operation.Produces = append(operation.Produces, "text/plain")
		case "html", "text/html":
			operation.Produces = append(operation.Produces, "text/html")
		case "mpfd", "multipart/form-data":
			operation.Produces = append(operation.Produces, "multipart/form-data")
		default:
			operation.Produces = append(operation.Produces, "*/*")
		}
	}
	return nil
}

func (operation *Operation) ParseRouterComment(commentLine string) error {
	re := regexp.MustCompile(`([\w\.\/\-{}]+)[^\[]+\[([^\]]+)`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 3 {
		return fmt.Errorf("Can not parse router comment \"%s\", skipped.", commentLine)
	}
	path := matches[1]
	httpMethod := matches[2]

	operation.Path = path
	operation.HttpMethod = strings.ToUpper(httpMethod)

	return nil
}

// @Success 200 {string} string model.File "ok"
func (operation *Operation) ParseResponseComment(commentLine string) error {
	re := regexp.MustCompile(`([\d]+)[\s]+([\w\{\}]+)[\s]+([@\w\-\.\/]+)[^"]*(.*)?`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 5 {
		return fmt.Errorf("Can not parse response comment \"%s\".", commentLine)
	}

	response := spec.Response{}

	code, _ := strconv.Atoi(matches[1])

	response.Description = strings.Trim(matches[4], "\"")

	resType := strings.Trim(matches[2], "{}")
	dataType := matches[3]
	if operation.parser != nil { // checking refType has existing in 'TypeDefinitions'
		if strings.Index(strings.TrimSpace(dataType), "@") == 0 {
			dataType = dataType[1:]
			if ref, ok := operation.parser.Definitions[dataType]; ok {
				operation.parser.registerTypes[dataType] = ref
			}
		}
	}

	// so we have to know all type in app
	//TODO: we might omitted schema.type if schemaType equals 'object'
	response.Schema = &spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"object"}}}
	log.Println(resType)
	if resType == "object" {
		response.Schema.Ref = spec.Ref{
			Ref: jsonreference.MustCreateRef("#/definitions/" + dataType),
		}
		response.Schema.Type = []string{"object"}
	}

	if resType == "array" {
		response.Schema.Items = &spec.SchemaOrArray{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/" + dataType)},
				},
			},
		}
		response.Schema.Type = []string{"array"}

	}

	if operation.Responses == nil {
		operation.Responses = &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: make(map[int]spec.Response),
			},
		}
	}

	operation.Responses.StatusCodeResponses[code] = response

	return nil
}

func (operation *Operation) ParseEmptyResponseComment(commentLine string) error {
	re := regexp.MustCompile(`([\d]+)[\s]+"(.*)"`)
	var matches []string

	if matches = re.FindStringSubmatch(commentLine); len(matches) != 3 {
		return fmt.Errorf("can not parse empty response comment \"%s\"", commentLine)
	}

	response := spec.Response{}

	code, _ := strconv.Atoi(matches[1])

	response.Description = strings.Trim(matches[2], "")

	if operation.Responses == nil {
		operation.Responses = &spec.Responses{
			ResponsesProps: spec.ResponsesProps{
				StatusCodeResponses: make(map[int]spec.Response),
			},
		}
	}

	operation.Responses.StatusCodeResponses[code] = response

	return nil
}

// createParamter returns swagger spec.Parameter for gived  paramType, description, paramName, schemaType, required
func createParameter(paramType, description, paramName, schemaType string, required bool) spec.Parameter {
	// //five possible parameter types. 	query, path, body, header, form
	paramProps := spec.ParamProps{
		Name:        paramName,
		Description: description,
		Required:    required,
		In:          paramType,
	}
	if paramType == "body" {
		paramProps.Schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{schemaType},
			},
		}
		parameter := spec.Parameter{
			ParamProps: paramProps,
		}

		return parameter
	} else {
		parameter := spec.Parameter{
			ParamProps: paramProps,
			SimpleSchema: spec.SimpleSchema{
				Type: schemaType,
			},
		}
		return parameter

	}
}

func createFormDataParameter(paramType, description, paramName, schemaType string, required bool) spec.Parameter {
	// //five possible parameter types. 	query, path, body, header, form
	paramProps := spec.ParamProps{
		Name:        paramName,
		Description: description,
		Required:    required,
		In:          paramType,
	}
	if paramType == "body" {
		paramProps.Schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{schemaType},
			},
		}
		parameter := spec.Parameter{
			ParamProps: paramProps,
		}

		return parameter
	} else {
		parameter := spec.Parameter{
			ParamProps: paramProps,
			SimpleSchema: spec.SimpleSchema{
				Type: schemaType,
			},
		}
		return parameter

	}
}

// parseTag
func (operation *Operation) ParseTagComment(commentLine string) []string {
	attr := strings.Split(commentLine, " ")
	var r = []string{}
	if len(attr) > 1 {
		r = append(r, attr[1])
	}
	if len(attr) > 2 {
		r = append(r, attr[2])
	}
	return r
}

// getPropertyName returns the string value for the given field if it exists, otherwise it panics.
// allowedValues: array, boolean, integer, null, number, object, string
func getPropertyName(field *ast.Field) string {
	var shouldTransInt = []string{"int64", "uint64", "int32", "uint32", "int8", "uint8"}
	var shouldTransFloat = []string{"float64", "float32", "float8"}
	var name string
	if _, ok := field.Type.(*ast.SelectorExpr); ok {
		// Support for time.Time as a structure field
		return "string"

	} else if astTypeIdent, ok := field.Type.(*ast.Ident); ok {
		name = astTypeIdent.Name
		for _, s := range shouldTransInt {
			if s == name {
				name = "number"
			}
		}
		for _, s := range shouldTransFloat {
			if s == name {
				name = "number"
			}
		}
	} else if _, ok := field.Type.(*ast.StarExpr); ok {
		panic("not supported astStarExpr yet.")
	} else if _, ok := field.Type.(*ast.MapType); ok { // if map
		return "object"
	} else if _, ok := field.Type.(*ast.ArrayType); ok { // if array
		return "array"
	} else if _, ok := field.Type.(*ast.StructType); ok { // if struct
		return "object"
	} else {
		log.Fatalf("Something goes wrong: %#v", field.Type)
	}

	return name
}

func snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}
