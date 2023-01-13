package template

const (
	//Trans define a trans method
	Trans = `
func (m *default{{.upperStartCamelObject}}Model) Trans(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return fn(ctx, session)
	})
}
`
	//TransMethod define a trans method
	TransMethod = `Trans(ctx context.Context, fn func(ctx context.Context, session sqlx.Session)) error`
)
