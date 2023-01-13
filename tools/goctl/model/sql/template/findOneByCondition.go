package template

const (
	FindOneByCondition = `
func (m *default{{.upperStartCamelObject}}Model) FindOneByCondition(ctx context.Context, builder squirrel.SelectBuilder) (*{{.upperStartCamelObject}},error) {
	query, values, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	var resp {{.upperStartCamelObject}}
	if err = m.conn.QueryRowPartialCtx(ctx, &resp, query, values...); err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
`
	FindOneByConditionMethod = `func (m *default{{.upperStartCamelObject}}Model) FindOneByCondition(ctx context.Context, builder squirrel.SelectBuilder) (*{{.upperStartCamelObject}},error) `
)
