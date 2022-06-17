package gogen

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/util"
)

const (
	FILE = "marshal.tpl"
)

var marshalTemplate string

func GenMarshal(api *spec.ApiSpec, category string) (string, error) {
	types := api.Types
	need2gen := []spec.Type{}
	//Filter out unnecessary generation types
	for _, tp := range types {
		name := tp.Name()
		if !isStartWith([]string{"Update", "Query", "Create"}, name) {
			need2gen = append(need2gen, tp)
		}
	}
	if len(need2gen) == 0 {
		return "", errors.New("no marshal and unMarsha func() to generate")
	}
	var builder strings.Builder
	for _, tp := range need2gen {
		var temp strings.Builder
		tableName := util.Title(tp.Name())
		temp.WriteString(fmt.Sprintf("// %s\n", tableName))
		structType, ok := tp.(spec.DefineStruct)
		if !ok {
			return "", fmt.Errorf("unspport struct type: %s", tp.Name())
		}
		marshal, err := buildMarshalFieldWrite(structType)
		if err != nil {
			return "", err
		}
		unMarshal, err := buildUnmarshalFieldWrite(structType)
		if err != nil {
			return "", err
		}
		dir, err := pathx.GetTemplateDir(category)
		if err != nil {
			return "", err
		}
		file := filepath.Join(dir, FILE)
		f, err := os.Create(file)
		if err != nil {
			return "", err
		}
		_, err = f.WriteString(marshalTemplate)
		if err != nil {
			return "", err
		}
		content, err := pathx.LoadTemplate(category, FILE, marshalTemplate)
		if err != nil {
			return "", err
		}
		data := map[string]interface{}{
			"upperStartCamelObject": tableName,
			"unmarshallFields":      unMarshal,
			"marshalFields":         marshal,
		}
		t := template.Must(template.New("marshalTemplate").Parse(content))
		buffer := new(bytes.Buffer)
		err = t.Execute(buffer, data)
		if err != nil {
			return "", err
		}
		code := formatCode(buffer.String())
		temp.WriteString(code)
		temp.WriteString("\n\n")
		builder.WriteString(temp.String())
	}
	return builder.String(), nil
}

func buildMarshalFieldWrite(tp spec.DefineStruct) (string, error) {
	var builder strings.Builder
	writeMarshalField(&builder, tp)
	return builder.String(), nil
}

func writeMarshalField(writer io.Writer, tp spec.DefineStruct) error {
	for _, member := range tp.Members {
		fmt.Fprintf(writer, "\tr.%s = p.%s\n", member.Name, member.Name)
	}
	return nil
}

func buildUnmarshalFieldWrite(tp spec.DefineStruct) (string, error) {
	var builder strings.Builder
	writeUmMarshalField(&builder, tp)
	return builder.String(), nil

}
func writeUmMarshalField(writer io.Writer, tp spec.DefineStruct) error {
	for _, member := range tp.Members {
		fmt.Fprintf(writer, "\tp.%s = s.%s \n", member.Name, member.Name)
	}
	return nil
}
func isStartWith(arr []string, target string) bool {
	for _, value := range arr {
		if strings.Contains(target, value) {
			return true
		}
	}
	return false
}