// source:tools/goctl/api/gogen/svc.tpl
package svc

import (
	{{.configImport}}
	{{.rpcImport}}
)

type ServiceContext struct {
	Config {{.config}}
	{{.middleware}}
	{{.rpc}}Cli
}

func NewServiceContext(c {{.config}}) *ServiceContext {
	return &ServiceContext{
		Config: c,
		{{.middlewareAssignment}}
		{{.rpcInit}}
	}
}
