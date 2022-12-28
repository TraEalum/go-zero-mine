package template

const (
	// Lists defines a template for lists code in model
	Lists = `
func (m *default{{.upperStartCamelObject}}Model) FindList(ctx context.Context, selectBuilder squirrel.SelectBuilder, totalCount ...*int64) (*[]{{.upperStartCamelObject}}, error) {
	var resp []{{.upperStartCamelObject}}

	query, values, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}
	
	if err = m.conn.QueryRowsCtx(ctx, &resp, query, values...);err != nil{
		return nil, err
	}

	if len(totalCount) != 0 {
		count := struct{Count int64 {{.countTag}}}{}
		query, values, err =sqlBuilder.Delete(selectBuilder, "Columns").(squirrel.SelectBuilder).Columns("COUNT({{.camel2Snake}}.{{.primaryKey}}) as count").RemoveOffset().ToSql()
		if err = m.conn.QueryRowCtx(ctx, &count, query, values...);err != nil {
			return nil, err
		}

		*totalCount[0] = count.Count
	}

	return &resp, nil
}
`

	// ListsMethod defines an interface method template for lists code in model
	ListsMethod = `FindList(ctx context.Context, selectBuilder squirrel.SelectBuilder, totalCount ...*int64) (*[]{{.upperStartCamelObject}},error)`
)
