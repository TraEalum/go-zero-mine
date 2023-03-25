package generator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

const old = `
// Package svc
// source: tools/goctl/rpc/generator/svc.tpl
package svc

import (
	"go-service/app/test/model"
	"go-service/app/test/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config config.Config
	// <codeGeneratedModelDefine>
	AntAuditOrderModel model.AntAuditOrderModel

	// </codeGeneratedModelDefine>
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.DB.DataSource)

	return &ServiceContext{
		Config: c,
		// <codeGeneratedModelInit>
		AntAuditOrderModel: model.NewAntAuditOrderModel(sqlConn),

		// </codeGeneratedModelInit>
	}
}
`
const new = `
// Package svc
// source: tools/goctl/rpc/generator/svc.tpl
package svc

import (
	"go-service/app/test/model"
	"go-service/app/test/rpc/internal/config"

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
		Config: c,
		// <codeGeneratedModelInit>
		{{.modelInit}}
		// </codeGeneratedModelInit>
	}
}
`

func Test_replaceTags1(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "test",
			args:    args{content: []byte(old)},
			want:    new,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := replaceTags(tt.args.content)
			if !tt.wantErr(t, err, fmt.Sprintf("replaceTags(%v)", tt.args.content)) {
				return
			}
			assert.Equalf(t, tt.want, got, "replaceTags(%v)", tt.args.content)
		})
	}
}
