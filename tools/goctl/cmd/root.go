package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"
	"text/template"

	"github.com/zeromicro/go-zero/tools/goctl/apigen"
	"github.com/zeromicro/go-zero/tools/goctl/proto"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
	"github.com/withfig/autocomplete-tools/integrations/cobra"
	"github.com/zeromicro/go-zero/tools/goctl/api"
	"github.com/zeromicro/go-zero/tools/goctl/bug"
	"github.com/zeromicro/go-zero/tools/goctl/docker"
	"github.com/zeromicro/go-zero/tools/goctl/env"
	"github.com/zeromicro/go-zero/tools/goctl/internal/version"
	"github.com/zeromicro/go-zero/tools/goctl/kube"
	"github.com/zeromicro/go-zero/tools/goctl/migrate"
	"github.com/zeromicro/go-zero/tools/goctl/model"
	"github.com/zeromicro/go-zero/tools/goctl/quickstart"
	"github.com/zeromicro/go-zero/tools/goctl/rpc"
	"github.com/zeromicro/go-zero/tools/goctl/tpl"
	"github.com/zeromicro/go-zero/tools/goctl/upgrade"
)

const (
	codeFailure = 1
	dash        = "-"
	doubleDash  = "--"
	assign      = "="
)

var (
	//go:embed usage.tpl
	usageTpl string

	rootCmd = &cobra.Command{
		Use:   "goctl",
		Short: "A cli tool to generate go-zero code",
		Long: "A cli tool to generate api, zrpc, model code\n\n" +
			"GitHub: https://github.com/zeromicro/go-zero\n" +
			"Site:   https://go-zero.dev",
	}
)

// Execute executes the given command
func Execute() {
	//	fmt.Printf("%#v", os.Args) // type  args
	os.Args = supportGoStdFlag(os.Args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(aurora.Red(err.Error()))
		os.Exit(codeFailure)
	}
}

func supportGoStdFlag(args []string) []string {
	copyArgs := append([]string(nil), args...)
	parentCmd, _, err := rootCmd.Traverse(args[:1])
	if err != nil { // ignore it to let cobra handle the error.
		return copyArgs
	}

	for idx, arg := range copyArgs[0:] {
		parentCmd, _, err = parentCmd.Traverse([]string{arg})
		if err != nil { // ignore it to let cobra handle the error.
			break
		}
		if !strings.HasPrefix(arg, dash) {
			continue
		}

		flagExpr := strings.TrimPrefix(arg, doubleDash)
		flagExpr = strings.TrimPrefix(flagExpr, dash)
		flagName, flagValue := flagExpr, ""
		assignIndex := strings.Index(flagExpr, assign)
		if assignIndex > 0 {
			flagName = flagExpr[:assignIndex]
			flagValue = flagExpr[assignIndex:]
		}

		if !isBuiltin(flagName) {
			// The method Flag can only match the user custom flags.
			f := parentCmd.Flag(flagName)
			if f == nil {
				continue
			}
			if f.Shorthand == flagName {
				continue
			}
		}

		goStyleFlag := doubleDash + flagName
		if assignIndex > 0 {
			goStyleFlag += flagValue
		}

		copyArgs[idx] = goStyleFlag
	}
	return copyArgs
}

func isBuiltin(name string) bool {
	return name == "version" || name == "help"
}

func init() {
	cobra.AddTemplateFuncs(template.FuncMap{
		"blue":    blue,
		"green":   green,
		"rpadx":   rpadx,
		"rainbow": rainbow,
	})

	rootCmd.Version = fmt.Sprintf(
		"%s %s/%s", version.BuildVersion,
		runtime.GOOS, runtime.GOARCH)

	rootCmd.SetUsageTemplate(usageTpl)
	rootCmd.AddCommand(api.Cmd)
	rootCmd.AddCommand(bug.Cmd)
	rootCmd.AddCommand(docker.Cmd)
	rootCmd.AddCommand(kube.Cmd)
	rootCmd.AddCommand(env.Cmd)
	rootCmd.AddCommand(model.Cmd)
	rootCmd.AddCommand(migrate.Cmd)
	rootCmd.AddCommand(quickstart.Cmd)
	rootCmd.AddCommand(rpc.Cmd)
	rootCmd.AddCommand(tpl.Cmd)
	rootCmd.AddCommand(upgrade.Cmd)
	rootCmd.AddCommand(cobracompletefig.CreateCompletionSpecCommand())
	rootCmd.AddCommand(proto.Cmd)
	rootCmd.AddCommand(apigen.Cmd)
}
