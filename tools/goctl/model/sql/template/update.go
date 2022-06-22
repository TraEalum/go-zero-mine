package template

const (
	// Update defines a template for generating update codes
	Update = `
func (m *default{{.upperStartCamelObject}}Model) Update(ctx context.Context, session sqlx.Session, updateBuilder squirrel.UpdateBuilder) error {
	var err error
	query, _, err := updateBuilder.ToSql()
	if err != nil {
		return err
	}

	{{if .withCache}}{{.keys}}
    _, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		if session != nil {
			_, err = session.ExecCtx(ctx, query)
			return err
		}
		return conn.ExecCtx(ctx, query)
	}, {{.keyValues}}){{else}}
	if session != nil {
		_, err = session.ExecCtx(ctx, query)
		return err
	}
	
    _, err = m.conn.ExecCtx(ctx, query){{end}}

	return err
}
`

	// UpdateMethod defines an interface method template for generating update codes
	UpdateMethod = `Update(ctx context.Context, session sqlx.Session, updateBuilder squirrel.UpdateBuilder) error`
)
