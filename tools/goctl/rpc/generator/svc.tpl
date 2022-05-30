package svc

import (
	{{.imports}}
	"go-service/app/{{.serviceName}}/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	)

type ServiceContext struct {
	Config config.Config

	{{.modelDefine}}
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config:c,

		{{.modelInit}}
	}
}
 