package main

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/load"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/tools/goctl/cmd"
	"github.com/zeromicro/go-zero/tools/goctl/internal/version"
	"log"
)

func main() {
	fmt.Println("fmt-goctl -v:", version.GetGoctlVersion())
	log.Println("log-goctl -v:", version.GetGoctlVersion())
	logx.Disable()
	load.Disable()
	cmd.Execute()
}
