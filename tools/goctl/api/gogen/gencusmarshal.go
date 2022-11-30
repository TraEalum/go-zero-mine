package gogen

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/stringx"
	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	"github.com/zeromicro/go-zero/tools/goctl/util"

	"io"
	"os"
	"path"
	"strings"
	"text/template"

	util2 "github.com/zeromicro/go-zero/tools/goctl/api/util"
)


//go:embed customize_marshal.tpl
var customizeMarshalTemplate string

func GenCustomizeMarshal(api *spec.ApiSpec, category string) error {
	types := api.Types
	need2gen := []spec.Type{}
	serviceName := api.Service.Name
	//获取fields
	fields := getFields(api)
	if len(fields) == 0 {
		return errors.New("can not read fields")
	}

	//Filter out unnecessary generation types
	for _, tp := range types {
		name := tp.Name()
		if stringx.Contains(fields, name) {
			need2gen = append(need2gen, tp)
		}
	}

	if len(need2gen) == 0 {
		return errors.New("no marshal and unMarsha func() to generate")
	}


	for _, tp := range need2gen {
		var temp strings.Builder
		tableName := util.Title(tp.Name())

		structType, ok := tp.(spec.DefineStruct)
		if !ok {
			return fmt.Errorf("unspport struct type: %s", tp.Name())
		}

		marshal, err := buildMarshalFieldWrite(structType)
		if err != nil {
			return err
		}

		unMarshal, err := buildUnmarshalFieldWrite(structType)
		if err != nil {
			return err
		}

		data := map[string]interface{}{
			"upperStartCamelObject": tableName,
			"unmarshallFields":      unMarshal,
			"marshalFields":         marshal,
			"importProto":           fmt.Sprintf("import \"go-service/app/%s/rpc/proto\"", serviceName),
		}

		t := template.Must(template.New("customizeMarshalTemplate").Parse(customizeMarshalTemplate))
		buffer := new(bytes.Buffer)
		err = t.Execute(buffer, data)
		if err != nil {
			return err
		}

		// 写入文件
		fileName := fmt.Sprintf("%sType.go", tableName)
		filePath := path.Join(category, typesDir, fileName)

		context, _ := getCustomizationContext(filePath)
		buffer.WriteString(context)

		os.Remove(filePath)

		fp, _, err := util2.MaybeCreateFile(category, typesDir, fileName)

		defer fp.Close()

		if err != nil {
			return err
		}

		if fp == nil {
			fp, err = os.OpenFile(filePath, os.O_RDWR, 0666)
			if err != nil {
				return err
			}

		}

		code := formatCode(buffer.String())
		temp.WriteString(code)
		_, err = fp.WriteString(temp.String())

		if err != nil {
			return err
		}
	}

	return nil
}


// 获取自定义结构体
//获取表名
func getFields(api *spec.ApiSpec) []string {
	var res []string

	//读取多个导入的文件
	if len(api.Imports) > 0 {
		for _, v := range api.Imports {
			fileName := v.Value
			name := strings.ReplaceAll(fileName, `"`, "")
			tables, err := readFileField(name)
			if err != nil {
				continue
			}

			if len(tables) != 0 {
				res = append(res, tables...)
			}
		}
	}

	return res
}

// 读取参数文件， 获取自定义字段名称
func readFileField(path string) ([]string, error) {
	var res []string
	file, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	FileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if FileInfo.Size() < 0 {
		return nil, errors.New("file is empty")
	}

	bufR := bufio.NewReader(file)
	defer file.Close()

	for {
		line, err := bufR.ReadString('\n')
		if strings.Contains(line, "Proto Customize Type:") {
			break
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
	}

	for {
		line, err := bufR.ReadString('\n')
		if strings.Contains(line, "Customize Type End") {
			break
		}

		line = strings.ReplaceAll(line, "\n", "")
		if len(line) > 3 {
			res = append(res, line[3:])

		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

	}

	return res, nil
}
