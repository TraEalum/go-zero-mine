package {{.PkgName}}

import (
	"net/http"

	json "github.com/json-iterator/go"

	{{.ImportPackages}}
	"comm/httpm"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		{{if .HasRequest}}var req types.{{.RequestType}}
		if err := httpx.Parse(r, &req); err != nil {
			httpm.ParamErrorResult(w, err)
			return
		}
		reqJson, _ := json.Marshal(&req)
		logx.WithContext(r.Context()).Infof("req: %s", string(reqJson))

		{{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
		{{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
			httpm.HttpResult(r, w, err, resp)
	}
}
