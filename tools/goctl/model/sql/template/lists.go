package template

const (
	// Lists defines a template for lists code in model
	Lists = `
func (m *default{{.upperStartCamelObject}}Model) Lists(ctx context.Context, where map[interface{}]interface{}) (*[]{{.upperStartCamelObject}}, error) {
	query := fmt.Sprintf("select %s from %s limit 15", {{.lowerStartCamelObject}}Rows, m.table)
	var values []interface{}
	for column, value := range where {
		query += fmt.Sprintf(" and %s = ?", column)
		values = append(values, value)
	}
	var resp []{{.upperStartCamelObject}}
	err := m.conn.QueryRowCtx(ctx, &resp, query, values...)
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
	ListsMethod = `Lists(ctx context.Context, where map[interface{}]interface{}) (*[]{{.upperStartCamelObject}},error)`
)
