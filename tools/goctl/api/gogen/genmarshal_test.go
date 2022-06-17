package gogen

import (
	"fmt"
	"testing"

	"github.com/zeromicro/go-zero/tools/goctl/api/parser"
)

func TestGenModelCustom(t *testing.T) {
	fileName := "test.api"
	api, err := parser.Parse(fileName)
	if err != nil {
		fmt.Println(err.Error())
	}
	custom, err := GenMarshal(api, ".")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(custom)

}

func TestGenType(t *testing.T) {
	// dir :="."
	// cfg, err := config.NewConfig("")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fileName :="test.api"
	// api, err := parser.Parse(fileName)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	// err = genTypes(dir, cfg, api)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }
	GenTemplates()
	if content, ok := templates[typesTemplateFile]; ok {
		fmt.Println(content)
	}
}
