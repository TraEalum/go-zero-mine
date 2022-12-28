package gogen

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/api/spec"
	apiutil "github.com/zeromicro/go-zero/tools/goctl/api/util"
	"github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
)

const typesFile = "types"

//go:embed types.tpl
var typesTemplate string

// BuildTypes gen types to string
func BuildTypes(types []spec.Type, apiFile ...string) (string, error) {
	var tys = make(map[string]struct{})
	var err error
	var builder strings.Builder
	first := true

	if len(apiFile) > 0 {
		tys, err = gGetReqMessage(apiFile[0])
		if err != nil {
			return "", err
		}
	}



	for _, tp := range types {
		if first {
			first = false
		} else {
			builder.WriteString("\n\n")
		}

		if err := writeType(&builder, tp, tys); err != nil {
			return "", apiutil.WrapErr(err, "Type "+tp.Name()+" generate error")
		}
	}

	return builder.String(), nil
}

// 获取get方法的请求参数名称
func gGetReqMessage(path string) (map[string]struct{}, error) {
	var res = make(map[string]struct{})
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
		if strings.Contains(line, "service") {
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

	reg := `(\([^\)]*\))`
	compile := regexp.MustCompile(reg)

	for {
		line, err := bufR.ReadString('\n')
		if strings.Contains(line, "Service Record End") {
			break
		}

		line = strings.ReplaceAll(line, "\n", "")
		line = strings.TrimSpace(line)
		split := strings.Split(line, " ")

		if split[0] == "get" {

			s := compile.FindString(line)
			s = strings.ReplaceAll(s, "(", "")
			s = strings.ReplaceAll(s, ")", "")
			res[s] = struct{}{}
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

func genTypes(dir string, cfg *config.Config, api *spec.ApiSpec, marshalFlag, apiFile string) error {
	var sep = `\`
	if runtime.GOOS == "linux" {
		sep = "/"
	}

	split := strings.Split(apiFile, sep)

	importFile := ""

	for i := 0; i < len(split) -1 ; i++ {
		importFile = path.Join(importFile, split[i])
	}


	val, err := BuildTypes(api.Types, apiFile)
	if err != nil {
		return err
	}

	typeFilename, err := format.FileNamingFormat(cfg.NamingFormat, typesFile)
	if err != nil {
		return err
	}

	typeFilename = typeFilename + ".go"
	filename := path.Join(dir, typesDir, typeFilename)
	os.Remove(filename)

	go func() {
		err = GenMarshal(api, dir, importFile)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("generate marsha file error")
			return
		}
	}()

	go func() {
		err = GenCustomizeMarshal(api, dir, importFile)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("generate cusMarsha file error")
			return
		}
	}()


	if err != nil {
		return err
	}

	return genFile(fileGenConfig{
		dir:             dir,
		subdir:          typesDir,
		filename:        typeFilename,
		templateName:    "typesTemplate",
		category:        category,
		templateFile:    typesTemplateFile,
		builtinTemplate: typesTemplate,
		data: map[string]interface{}{
			"types":        val,
			"containsTime": false,
		},
	})
}

func writeType(writer io.Writer, tp spec.Type, tys map[string]struct{}) error {
	structType, ok := tp.(spec.DefineStruct)
	if !ok {
		return fmt.Errorf("unspport struct type: %s", tp.Name())
	}

	fmt.Fprintf(writer, "type %s struct {\n", util.Title(tp.Name()))
	for _, member := range structType.Members {

		if _, ok := tys[tp.Name()]; ok {
			t1 := strings.Replace(member.Tag, "\"`", ",optional\"`", -1)
			member.Tag = strings.Replace(t1, "json", "form", -1)
			}


		if member.IsInline {
			if _, err := fmt.Fprintf(writer, "%s\n", strings.Title(member.Type.Name())); err != nil {
				return err
			}

			continue
		}

		if err := writeProperty(writer, member.Name, member.Tag, member.GetComment(), member.Type, 1); err != nil {
			return err
		}
	}
	fmt.Fprintf(writer, "}")
	return nil
}
