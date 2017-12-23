package swag

import (
	"go/ast"
	"log"
)

// getPropertyName returns the string value for the given field if it exists, otherwise it panics.
// allowedValues: array, boolean, integer, null, number, object, string
func getPropertyName(field *ast.Field) string {
	var shouldTransInt = []string{"int64","uint64","int32","uint32","int8","uint8"}
	var shouldTransFloat = []string{"float64","float32","float8"}
	var name string
	if astTypeSelectorExpr, ok := field.Type.(*ast.SelectorExpr); ok {
		// Support for time.Time as a structure field
		if "Time" == astTypeSelectorExpr.Sel.Name {
			return "string"
		}
		if "ObjectId" ==  astTypeSelectorExpr.Sel.Name {
			return "string"
		}
		panic("not supported 'astSelectorExpr' yet.")

	} else if astTypeIdent, ok := field.Type.(*ast.Ident); ok {
		name = astTypeIdent.Name
		for _ , s := range shouldTransInt {
			if s == name {
				name = "number"
			}
		}
		for _ , s := range shouldTransFloat {
			if s == name {
				name = "number"
			}
		}
	} else if _, ok := field.Type.(*ast.StarExpr); ok {
		panic("not supported astStarExpr yet.")
	} else if _, ok := field.Type.(*ast.MapType); ok { // if map
		//TODO: support map
		return "object"
	} else if _, ok := field.Type.(*ast.ArrayType); ok { // if array
		return "array"
	} else if _, ok := field.Type.(*ast.StructType); ok { // if struct
		//TODO: support nested struct
		return "object"
	} else {
		log.Fatalf("Something goes wrong: %#v", field.Type)
	}

	return name
}
