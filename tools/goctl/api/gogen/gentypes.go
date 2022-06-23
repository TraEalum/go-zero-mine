package gogen

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path"
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
func BuildTypes(types []spec.Type) (string, error) {
	var builder strings.Builder
	first := true
	for _, tp := range types {
		if first {
			first = false
		} else {
			builder.WriteString("\n\n")
		}
		if err := writeType(&builder, tp); err != nil {
			return "", apiutil.WrapErr(err, "Type "+tp.Name()+" generate error")
		}
	}

	return builder.String(), nil
}

func genTypes(dir string, cfg *config.Config, api *spec.ApiSpec, marshalFlag string) error {
	val, err := BuildTypes(api.Types)
	if err != nil {
		return err
	}

	typeFilename, err := format.FileNamingFormat(cfg.NamingFormat, typesFile)
	if err != nil {
		return err
	}
	protoPath := fmt.Sprintf("\"go-service/app/%s/rpc/proto\"", api.Service.Name)
	typeFilename = typeFilename + ".go"
	filename := path.Join(dir, typesDir, typeFilename)
	os.Remove(filename)
	marshal := ""
	if strings.ToLower(marshalFlag) == "yes" {
		marshal, err = GenMarshal(api, dir)
	}
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
			"porto":        protoPath,
			"marshal":      marshal,
		},
	})
}

func writeType(writer io.Writer, tp spec.Type) error {
	structType, ok := tp.(spec.DefineStruct)
	if !ok {
		return fmt.Errorf("unspport struct type: %s", tp.Name())
	}

	fmt.Fprintf(writer, "type %s struct {\n", util.Title(tp.Name()))
	for _, member := range structType.Members {
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
