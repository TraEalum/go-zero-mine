package main

import (
	"flag"
	"fmt"

	{{.importPackages}}
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

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
