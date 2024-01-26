package {{.PkgName}}

import (
	"net/http"

	json "github.com/json-iterator/go"

	{{.ImportPackages}}
	"comm/httpm"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			httpm.ParamErrorResultV2(w, err)

			return
		}

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
			httpm.HttpResultV2(r, w, err, resp{{if .IsList}}, true{{end}})
	}
}
