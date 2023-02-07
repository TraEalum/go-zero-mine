package main

import (
	"flag"
	"fmt"

	{{.imports}}

	"comm/configm"
	"comm/util"
	
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	if configFile != nil && *configFile != "" {
		conf.MustLoad(*configFile, &c)
	} else {
		configm.LoadConfig(configm.ConfigInfo{
			ServerType: "rpc",
			Server:     "{{.serviceKey}}",
		}, &c)
	}

	ctx := svc.NewServiceContext(c)
	svr := server.New{{.serviceNew}}Server(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		{{.pkg}}.Register{{.service}}Server(grpcServer, svr)

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	s.AddUnaryInterceptors(util.LoggerInterceptor)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
