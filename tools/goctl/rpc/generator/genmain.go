package generator

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	conf "github.com/zeromicro/go-zero/tools/goctl/config"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
	"github.com/zeromicro/go-zero/tools/goctl/util"
	"github.com/zeromicro/go-zero/tools/goctl/util/format"
	"github.com/zeromicro/go-zero/tools/goctl/util/pathx"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

//go:embed main.tpl
var mainTemplate string

type MainServiceTemplateData struct {
	Service   string
	ServerPkg string
	Pkg       string
}

// GenMain generates the main file of the rpc service, which is an rpc service program call entry
func (g *Generator) GenMain(ctx DirContext, proto parser.Proto, cfg *conf.Config,
	c *ZRpcContext) error {
	mainFilename, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	fileName := filepath.Join(ctx.GetMain().Filename, fmt.Sprintf("%v.go", mainFilename))
	imports := make([]string, 0)
	//pbImport := fmt.Sprintf(`proto "proto/%s"`, proto.Service[0].Name)
	pbImport := fmt.Sprintf(`proto "proto/%s"`, proto.PbPackage)
	svcImport := fmt.Sprintf(`"%v"`, ctx.GetSvc().Package)
	configImport := fmt.Sprintf(`"%v"`, ctx.GetConfig().Package)
	imports = append(imports, configImport, pbImport, svcImport)

	var serviceNames []MainServiceTemplateData
	var registerServer string
	for _, e := range proto.Service {

		var (
			remoteImport string
			serverPkg    string
		)
		if !c.Multiple {
			serverPkg = "server"
			remoteImport = fmt.Sprintf(`"%v"`, ctx.GetServer().Package)
		} else {
			childPkg, err := ctx.GetServer().GetChildPackage(e.Name)
			if err != nil {
				return err
			}

			// serverPkg = filepath.Base(childPkg + "Server")
			serverPkg = stringx.From(e.Name).ToCamelWithStartLower() + "Server"
			remoteImport = fmt.Sprintf(`%s "%v"`, serverPkg, childPkg)
		}

		imports = append(imports, remoteImport)
		serviceNames = append(serviceNames, MainServiceTemplateData{
			Service:   parser.CamelCase(e.Name),
			ServerPkg: serverPkg,
			Pkg:       "proto",
		})

		registerServer += fmt.Sprintf("\t\tproto.Register%sServer(grpcServer, %s.New%sServer(ctx))\n", parser.CamelCase(e.Name), serverPkg, parser.CamelCase(e.Name))
	}

	// len大于二 只修改注册服务行代码
	if c.Multiple && len(proto.Service) > 1 {
		start := time.Now()
		fmt.Println("gen main方法-upDateNewServer耗时开始时间:", start)
		err2 := upDateNewServer(fileName, registerServer, imports)
		fmt.Println("gen main方法-upDateNewServer执行耗时:", time.Since(start))
		return err2
	}

	text, err := pathx.LoadTemplate(category, mainTemplateFile, mainTemplate)
	if err != nil {
		return err
	}

	etcFileName, err := format.FileNamingFormat(cfg.NamingFormat, ctx.GetServiceName().Source())
	if err != nil {
		return err
	}

	return util.With("main").GoFmt(true).Parse(text).SaveTo(map[string]interface{}{
		"serviceName":  etcFileName,
		"serviceNames": serviceNames,
		"imports":      strings.Join(imports, pathx.NL),
		"pkg":          "proto",
		"serviceNew":   stringx.From(proto.Service[0].Name).ToCamel(),
		"service":      parser.CamelCase(proto.Service[0].Name),
		"serviceKey":   proto.Service[0].Name,
	}, fileName, false)
}

func upDateNewServer(fileName, registerServer string, imports []string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := bufio.NewReader(f)
	newBuf := new(bytes.Buffer)
Loop:
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return errors.New("Read file error!")
			}
		}

		// if strings.Contains(line, "\"fmt\"") {
		// 	newBuf.WriteString(line)
		// 	newBuf.WriteString("\n")

		// 	for {
		// 		fmt.Println("upDateNewServer 1")
		// 		line, _ := buf.ReadString('\n')
		// 		if strings.Contains(line, "comm/configm") {
		// 			// 写入新imports
		// 			for _, v := range imports {
		// 				newBuf.WriteString("\t" + v + "\n")
		// 			}

		// 			newBuf.WriteString("\n")
		// 			newBuf.WriteString(line)

		// 			continue Loop
		// 		}
		// 	}
		// }

		if strings.Contains(line, "zrpc.MustNewServer") {
			newBuf.WriteString(line)

			for {
				line, _ := buf.ReadString('\n')
				if strings.Contains(line, "service.DevMode") {
					// 写入新注册服务
					newBuf.WriteString(registerServer)

					newBuf.WriteString("\n")
					newBuf.WriteString(line)

					continue Loop
				}
			}
		}

		newBuf.WriteString(line)
	}

	err = ioutil.WriteFile(fileName, newBuf.Bytes(), 0666)
	if err != nil {
		fmt.Printf("生成paramFile文件失败:%v", err.Error())
		return err
	}

	return nil
}
