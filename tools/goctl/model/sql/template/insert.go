package template

const (
	// Insert defines a template for insert code in model
	Insert = `
func (m *default{{.upperStartCamelObject}}Model) Insert(ctx context.Context, session sqlx.Session, data *{{.upperStartCamelObject}}) (sql.Result,error) {
	var err error
	var ret sql.Result
	util.Orm(data, util.OrmScopeInsert)

	{{if .withCache}}{{.keys}}
    ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values ({{.expression}})", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
		if session != nil {
			return session.ExecCtx(ctx, query, {{.expressionValues}})
		}
		return conn.ExecCtx(ctx, query, {{.expressionValues}})
	}, {{.keyValues}}){{else}}query := fmt.Sprintf("insert into %s (%s) values ({{.expression}})", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
	if session != nil {
		ret, err = session.ExecCtx(ctx, query, {{.expressionValues}})
		return ret, err
	}

    ret,err =m.conn.ExecCtx(ctx, query, {{.expressionValues}}){{end}}
	return ret,err
}
`

	// InsertMethod defines an interface method template for insert code in model
	InsertMethod = `Insert(ctx context.Context, session sqlx.Session, data *{{.upperStartCamelObject}}) (sql.Result,error)`
)
