package gen

import (
	"bufio"
	"fmt"
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/parser"
	"github.com/zeromicro/go-zero/tools/goctl/model/sql/template"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func genTypes(table Table, methods string, withCache bool, dir, namingFormat string) (string, error) {
	fields := table.Fields
	var err error

	fields, err = getFields(fields, dir, namingFormat, table.Name)
	if err != nil {
		log.Fatalf("updateGen err: [%v]", err)

		return "", err
	}

	fieldsString, err := genFields(table, fields)
	if err != nil {
		return "", err
	}

	text, err := pathx.LoadTemplate(category, typesTemplateFile, template.Types)
	if err != nil {
		return "", err
	}

	output, err := util.With("types").
		Parse(text).
		Execute(map[string]interface{}{
			"withCache":             withCache,
			"method":                methods,
			"upperStartCamelObject": table.Name.ToCamel(),
			"lowerStartCamelObject": stringx.From(table.Name.ToCamel()).Untitle(),
			"fields":                fieldsString,
			"data":                  table,
		})
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func getFields(fields []*parser.Field, dir, namingFormat string, table stringx.String) ([]*parser.Field, error) {
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	tn := stringx.From(table.Source())
	modelFilename, err := format.FileNamingFormat(namingFormat,
		fmt.Sprintf("%s_model", tn.Source()))
	if err != nil {
		return nil, err
	}

	name := util.SafeString(modelFilename) + "_gen.go"
	filename := filepath.Join(dirAbs, name)

	if pathx.FileExists(filename) {
		file, err := os.OpenFile(filename, os.O_RDWR, 0666)
		if err != nil {
			log.Printf("Open file error!%v\n", err)
			return nil, err
		}
		defer file.Close()

		buf := bufio.NewReader(file)

		for {
			line, err := buf.ReadString('\n')

			line = strings.Replace(line, " ", "", -1)

			if strings.Contains(line, fmt.Sprintf("%s%s", table.ToCamel(), "struct")) {
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

		var oldFields = make([]*parser.Field, 0)
		var oldFieldMap = make(map[string]int)

		for {
			line, err := buf.ReadString('\n')
			if strings.Contains(line, "}") {
				break
			}

			tmpField := &parser.Field{}

			comment := strings.Split(line, "//")
			if len(comment) > 1 {
				c := strings.TrimSpace(strings.Replace(fmt.Sprintf("%s", comment[1]), "\n", "", -1))

				tmpField.Comment = c
			}

			split := strings.Split(comment[0], " ")
			cList := make([]string, 0)

			for _, v := range split {
				if v == "" {
					continue
				}

				cList = append(cList, strings.TrimSpace(v))
			}

			if len(cList) < 2 {
				continue
			}

			tmpField.Name = stringx.From(stringx.From(cList[0]).ToSnake())
			tmpField.DataType = strings.TrimSpace(strings.Replace(cList[1], "*", "", -1))

			if len(cList) > 2 {
				tmpField.NameOriginal = strings.TrimSpace(stringx.From(cList[0]).ToSnake())
			}

			oldFields = append(oldFields, tmpField)
			oldFieldMap[tmpField.Name.Source()] = 0

			if err != nil {
				if err == io.EOF {
					break
				} else {
					return nil, err
				}
			}
		}

		// 对比
		for _, v := range fields {
			_, ok := oldFieldMap[v.Name.Source()]
			if ok {
				oldFieldMap[v.Name.Source()] = 1
			} else {
				oldFields = append(oldFields, v)
			}
		}

		fields = make([]*parser.Field, 0)

		for _, v := range oldFields {
			value, ok := oldFieldMap[v.Name.Source()]
			if ok && value == 0 {
				continue
			}

			fields = append(fields, &parser.Field{
				NameOriginal:    v.NameOriginal,
				Name:            v.Name,
				DataType:        v.DataType,
				Comment:         v.Comment,
				SeqInIndex:      1,
				OrdinalPosition: 1,
			})
		}
	}

	return fields, nil
}
