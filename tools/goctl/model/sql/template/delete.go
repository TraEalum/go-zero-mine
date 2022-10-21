package template

const (
	// Delete defines a delete template
	Delete = `
func (m *default{{.upperStartCamelObject}}Model) Delete(ctx context.Context, session sqlx.Session, data *{{.upperStartCamelObject}}) error {
	query := util.Orm(data, util.OrmScopeDelete)
	var err error
	
	{{if .withCache}}{{if .containsIndexCache}}data, err:=m.FindOne(ctx, data.{{.primaryKey}})
	if err!=nil{
		return err
	}

{{end}}	{{.keys}}
    _, err {{if .containsIndexCache}}={{else}}:={{end}} m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		if session != nil {
			_, err = session.ExecCtx(ctx, query, data.{{.primaryKey}})
			return err
		}
		return conn.ExecCtx(ctx, query, data.{{.primaryKey}})
	}, {{.keyValues}}){{else}}
		if session != nil {
			_, err = session.ExecCtx(ctx, query, data.{{.primaryKey}})
			return err
		}

		_,err = m.conn.ExecCtx(ctx, query, data.{{.primaryKey}}){{end}}

	return err
}
`

	// DeleteMethod defines a delete template for interface method
	DeleteMethod = `Delete(ctx context.Context, session sqlx.Session, data *{{.upperStartCamelObject}}) error`
)
