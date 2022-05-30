package template

const (
	// Lists defines a template for lists code in model
	Lists = `
func (m *default{{.upperStartCamelObject}}Model) FindList(ctx context.Context, where map[string]interface{}, selectBuilder squirrel.SelectBuilder) (*[]{{.upperStartCamelObject}}, error) {
	var resp []{{.upperStartCamelObject}}
	for column, value := range where {
		selectBuilder = selectBuilder.Where(squirrel.Eq{column: value})
	}
	query, _, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}
	err = m.conn.QueryRowCtx(ctx, &resp, query)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
`

	// ListsMethod defines an interface method template for lists code in model
	ListsMethod = `FindList(ctx context.Context, where map[string]interface{}, selectBuilder squirrel.SelectBuilder) (*[]{{.upperStartCamelObject}},error)`
)
