package svc

import (
	{{.configImport}}
	{{.rpcImport}}
)

type ServiceContext struct {
	Config {{.config}}
	{{.middleware}}
	{{.rpc}}
}

func NewServiceContext(c {{.config}}) *ServiceContext {
	return &ServiceContext{
		Config: c,
		{{.middlewareAssignment}}
		{{.rpcInit}}
	}
}
