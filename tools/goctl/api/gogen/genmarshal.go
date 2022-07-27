package gogen

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	util2 "github.com/zeromicro/go-zero/tools/goctl/api/util"
	"github.com/zeromicro/go-zero/tools/goctl/util"
)

//go:embed marshal.tpl
var marshalTemplate string

func GenMarshal(api *spec.ApiSpec, category string) error {
	types := api.Types
	need2gen := []spec.Type{}

	//获取table
	tables := getTables(api)
	if len(tables) == 0 {
		return errors.New("can not read table")
	}

	//Filter out unnecessary generation types
	for _, tp := range types {
		name := tp.Name()
		if isContain(tables, name) {
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
		}

		t := template.Must(template.New("marshalTemplate").Parse(marshalTemplate))
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
		fmt.Fprintf(writer, "\tp.%s = r.%s \n", member.Name, member.Name)
	}

	return nil
}

//获取表名
func getTables(api *spec.ApiSpec) []string {
	var res []string

	//读取多个导入的文件
	if len(api.Imports) > 0 {
		for _, v := range api.Imports {
			fileName := v.Value
			name := strings.ReplaceAll(fileName, `"`, "")
			tables, err := readFileTable(name)
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

//读取参数文件，获取里面的表名称
func readFileTable(path string) ([]string, error) {
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
		if strings.Contains(line, "Already Exist Table") {
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
		if strings.Contains(line, "Exist Table End") {
			break
		}

		line = strings.ReplaceAll(line, "\n", "")
		res = append(res, line[3:])

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

//获取原有自定义的内容，如果存在的情况下
func getCustomizationContext(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	stat, err := f.Stat()
	if err != nil {
		return "", err
	}

	//空文件
	if stat.Size() == 0 {
		return "", errors.New("empty file")
	}

	bufW := new(bytes.Buffer)
	bufR := bufio.NewReader(f)

	//找到结束标志
	for {
		line, err := bufR.ReadString('\n')
		if strings.Contains(line, "TheEndLine") {
			break
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "", err
			}
		}
	}

	bufW.WriteString("\n")

	//开始读取自定义的内容
	for {
		line, err := bufR.ReadString('\n')
		bufW.WriteString(line)

		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "", err
			}
		}
	}

	return bufW.String(), nil
}
