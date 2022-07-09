// Package svc
// source: tools/goctl/rpc/generator/svc.tpl
package svc

import (
	{{.imports}}
	"go-service/app/{{.serviceName}}/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	)

type ServiceContext struct {
	Config config.Config
    // <codeGeneratedModelDefine>
	{{.modelDefine}}
	// </codeGeneratedModelDefine>
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:c,
        // <codeGeneratedModelInit>
		{{.modelInit}}
		// </codeGeneratedModelInit>
	}
}
 