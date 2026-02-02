package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"
)

// StructInfo хранит информацию о структуре для генерации
type StructInfo struct {
	Name     string
	Package  string
	FilePath string
	Fields   []*FieldInfo
	HasReset bool
}

// FieldInfo хранит информацию о поле структуры
type FieldInfo struct {
	Name      string
	TypeExpr  ast.Expr
	TypeStr   string
	BaseType  string
	IsPointer bool
	IsSlice   bool
	IsMap     bool
	IsArray   bool
	IsStruct  bool
	HasReset  bool
}

func main() {
	// Получаем путь к проекту
	projectPath, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Проходимся по всем пакетам и генерируем Reset методы
	err = filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем не директории
		if !info.IsDir() {
			return nil
		}

		// Пропускаем скрытые директории
		if strings.HasPrefix(filepath.Base(path), ".") {
			return filepath.SkipDir
		}

		// Пропускаем директории vendor, node_modules и cmd/reset
		base := filepath.Base(path)
		if base == "vendor" || base == "node_modules" || (base == "reset" && strings.Contains(path, "cmd/reset")) {
			return filepath.SkipDir
		}

		// Пропускаем директории с тестами
		if base == "_test" {
			return filepath.SkipDir
		}

		// Проверяем, есть ли в директории Go файлы
		files, err := filepath.Glob(filepath.Join(path, "*.go"))
		if err != nil || len(files) == 0 {
			return nil
		}

		// Пропускаем, если есть только сгенерированные файлы
		hasNonGen := false
		for _, f := range files {
			if !strings.HasSuffix(f, "_gen.go") && !strings.HasSuffix(f, "_test.go") {
				hasNonGen = true
				break
			}
		}
		if !hasNonGen {
			return nil
		}

		// Генерируем для текущего пакета
		if err := generateForPackage(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating for package %s: %v\n", path, err)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking project: %v\n", err)
		os.Exit(1)
	}
}

// generateForPackage генерирует Reset методы для всех структур в пакете
func generateForPackage(pkgPath string) error {
	fset := token.NewFileSet()

	// Получаем список Go файлов (кроме _gen.go и _test.go)
	files, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
	if err != nil {
		return fmt.Errorf("error listing files: %w", err)
	}

	// Парсим все подходящие .go файлы
	var astFiles []*ast.File
	var pkgName string
	for _, file := range files {
		// Пропускаем тестовые и сгенерированные файлы
		if strings.HasSuffix(file, "_test.go") || strings.HasSuffix(file, "_gen.go") {
			continue
		}

		// Парсим файл
		src, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("error reading file %s: %w", file, err)
		}

		astFile, err := parser.ParseFile(fset, file, src, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("error parsing file %s: %w", file, err)
		}

		// Получаем имя пакета из первого файла
		if pkgName == "" {
			pkgName = astFile.Name.Name
		}

		astFiles = append(astFiles, astFile)
	}

	if len(astFiles) == 0 {
		return nil
	}

	// Создаем информацию о типах
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	// Проверяем типы
	conf := types.Config{}
	_, err = conf.Check(pkgName, fset, astFiles, info)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: type checking failed for %s: %v (continuing anyway)\n", pkgPath, err)
	}

	// Находим структуры с комментарием // generate:reset
	var structs []*StructInfo
	for _, file := range astFiles {
		filePath := fset.File(file.Pos()).Name()
		fileStructs := findStructsWithReset(fset, file, info, pkgName, filePath)
		structs = append(structs, fileStructs...)
	}

	if len(structs) == 0 {
		return nil
	}

	// Генерируем код для reset.gen.go
	code := generateResetCode(pkgName, structs)

	// Записываем в файл
	outputPath := filepath.Join(pkgPath, "reset.gen.go")
	return os.WriteFile(outputPath, []byte(code), 0644)
}

// findStructsWithReset находит структуры с комментарием // generate:reset
func findStructsWithReset(fset *token.FileSet, file *ast.File, info *types.Info, pkgName, filePath string) []*StructInfo {
	var structs []*StructInfo

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// Проверяем комментарий перед объявлением
		if !hasResetComment(genDecl.Doc) {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Собираем информацию о полях
			fields := collectFieldInfo(structType.Fields.List, info)

			// Проверяем, есть ли у самой структуры метод Reset
			hasReset := hasResetMethodForType(typeSpec.Name.Name)

			structs = append(structs, &StructInfo{
				Name:     typeSpec.Name.Name,
				Package:  pkgName,
				FilePath: filePath,
				Fields:   fields,
				HasReset: hasReset,
			})
		}
	}

	return structs
}

// collectFieldInfo собирает информацию о полях структуры
func collectFieldInfo(fieldList []*ast.Field, info *types.Info) []*FieldInfo {
	var fields []*FieldInfo

	for _, field := range fieldList {
		typeAndValue, ok := info.Types[field.Type]
		var fieldTypes []types.Type

		if ok {
			fieldTypes = []types.Type{typeAndValue.Type}
		}

		for _, name := range field.Names {
			fieldType := exprToString(field.Type)
			var t types.Type

			if len(fieldTypes) > 0 {
				t = fieldTypes[0]
				fieldType = typeToString(t)
			}

			fieldInfo := &FieldInfo{
				Name:     name.Name,
				TypeExpr: field.Type,
				TypeStr:  fieldType,
				BaseType: getBaseType(fieldType),
			}

			// Определяем характеристики типа
			if t != nil {
				fieldInfo.IsPointer = isPointerType(t)
				fieldInfo.IsSlice = isSliceType(t)
				fieldInfo.IsMap = isMapType(t)
				fieldInfo.IsArray = isArrayType(t)
				fieldInfo.IsStruct = isStructType(t)
				fieldInfo.HasReset = hasResetMethodForType(fieldType)
			}

			fields = append(fields, fieldInfo)
		}

		// Анонимные поля (вложенные структуры)
		if len(field.Names) == 0 {
			// Пропускаем анонимные поля для простоты
			continue
		}
	}

	return fields
}

// generateResetCode генерирует код для reset.gen.go
func generateResetCode(pkgName string, structs []*StructInfo) string {
	var buf bytes.Buffer

	// Генерируем заголовок файла
	buf.WriteString("// Code generated by reset generator; DO NOT EDIT.\n\n")
	buf.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// Генерируем методы для каждой структуры
	for _, st := range structs {
		generateResetMethod(&buf, st)
	}

	// Форматируем код
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting code: %v\n", err)
		return buf.String()
	}
	return string(formatted)
}

// TemplateFuncs - функции для использования в шаблонах
var TemplateFuncs = template.FuncMap{
	"isStringType":        isStringType,
	"isBoolType":          isBoolType,
	"isNumericType":       isNumericType,
	"getBaseType":         getBaseType,
	"getSliceElementType": getSliceElementType,
}

// resetTemplate - шаблон для генерации метода Reset
var resetTemplate = template.Must(template.New("reset").Funcs(TemplateFuncs).Parse(`// Reset resets the {{.Name}} fields to their zero values
func (x *{{.Name}}) Reset() {
	if x == nil {
		return
	}
{{- range .Fields}}
{{- if .Name}}
	{{if .IsPointer}}if x.{{.Name}} != nil {
		{{- if isStringType .BaseType}}*x.{{.Name}} = ""
		{{- else if isBoolType .BaseType}}*x.{{.Name}} = false
		{{- else if isNumericType .BaseType}}*x.{{.Name}} = 0
		{{- else if .HasReset}}x.{{.Name}}.Reset()
		{{- else}}*x.{{.Name}} = {{.BaseType}}{}
		{{- end}}
	}
	{{- else if .IsSlice}}if x.{{.Name}} != nil {
		x.{{.Name}} = x.{{.Name}}[:0]
	}
	{{- else if .IsMap}}if x.{{.Name}} != nil {
		clear(x.{{.Name}})
	}
	{{- else if .IsArray}}x.{{.Name}} = [len(x.{{.Name}})]{{getSliceElementType .TypeStr}}{}
	{{- else if isStringType .TypeStr}}x.{{.Name}} = ""
	{{- else if isBoolType .TypeStr}}x.{{.Name}} = false
	{{- else if isNumericType .TypeStr}}x.{{.Name}} = 0
	{{- else if .IsStruct}}{{if .HasReset}}	resetter, ok := x.{{.Name}}.(interface{ Reset() })
	if ok {
		resetter.Reset()
	}
	{{- else}}x.{{.Name}} = {{.TypeStr}}{}
	{{- end}}
	{{- else}}x.{{.Name}} = {{.TypeStr}}{}
	{{- end}}
{{- end}}
{{- end}}
}
`))

// generateResetMethod генерирует метод Reset для структуры
func generateResetMethod(buf *bytes.Buffer, st *StructInfo) {
	if err := resetTemplate.Execute(buf, st); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template for %s: %v\n", st.Name, err)
	}
	buf.WriteString("\n")
}

// hasResetComment проверяет наличие комментария // generate:reset
func hasResetComment(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}
	for _, comment := range doc.List {
		if strings.Contains(comment.Text, "generate:reset") {
			return true
		}
	}
	return false
}

// exprToString преобразует ast.Expr в строку
func exprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	format.Node(&buf, token.NewFileSet(), expr)
	return buf.String()
}

// typeToString преобразует types.Type в строку
func typeToString(t types.Type) string {
	switch v := t.(type) {
	case *types.Basic:
		return v.Name()
	case *types.Pointer:
		return "*" + typeToString(v.Elem())
	case *types.Slice:
		return "[]" + typeToString(v.Elem())
	case *types.Map:
		return fmt.Sprintf("map[%s]%s", typeToString(v.Key()), typeToString(v.Elem()))
	case *types.Array:
		return fmt.Sprintf("[%d]%s", v.Len(), typeToString(v.Elem()))
	case *types.Named:
		return v.Obj().Name()
	case *types.Struct:
		return "struct{}"
	default:
		return ""
	}
}

// isPointerType проверяет, является ли тип указателем
func isPointerType(t types.Type) bool {
	_, ok := t.(*types.Pointer)
	return ok
}

// isSliceType проверяет, является ли тип слайсом
func isSliceType(t types.Type) bool {
	_, ok := t.(*types.Slice)
	return ok
}

// isMapType проверяет, является ли тип мапой
func isMapType(t types.Type) bool {
	_, ok := t.(*types.Map)
	return ok
}

// isArrayType проверяет, является ли тип массивом
func isArrayType(t types.Type) bool {
	_, ok := t.(*types.Array)
	return ok
}

// isStructType проверяет, является ли тип структурой
func isStructType(t types.Type) bool {
	switch v := t.(type) {
	case *types.Struct:
		return true
	case *types.Named:
		_, ok := v.Underlying().(*types.Struct)
		return ok
	default:
		return false
	}
}

// hasResetMethodForType проверяет наличие метода Reset у типа
func hasResetMethodForType(_ string) bool {
	// Для простоты считаем, что метод есть если имя типа указано явно
	// В реальной реализации нужно смотреть в info.Types и проверять методы
	return false
}

// getBaseType получает базовый тип из указателя (например, "*int" -> "int")
func getBaseType(typeStr string) string {
	if strings.HasPrefix(typeStr, "*") {
		return typeStr[1:]
	}
	return typeStr
}

// isStringType проверяет, является ли тип строкой
func isStringType(typeStr string) bool {
	return typeStr == "string"
}

// isBoolType проверяет, является ли тип bool
func isBoolType(typeStr string) bool {
	return typeStr == "bool"
}

// isNumericType проверяет, является ли тип числовым
func isNumericType(typeStr string) bool {
	numericTypes := []string{"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "complex64", "complex128"}
	return slices.Contains(numericTypes, typeStr)
}

// getSliceElementType получает тип элемента слайса или массива
func getSliceElementType(typeStr string) string {
	if strings.HasPrefix(typeStr, "[]") {
		return typeStr[2:]
	}
	if strings.HasPrefix(typeStr, "[") {
		if idx := strings.Index(typeStr, "]"); idx > 0 {
			return typeStr[idx+1:]
		}
	}
	return ""
}
