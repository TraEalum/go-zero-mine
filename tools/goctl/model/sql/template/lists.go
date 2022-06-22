package template

const (
	// Lists defines a template for lists code in model
	Lists = `
func (m *default{{.upperStartCamelObject}}Model) FindList(ctx context.Context, selectBuilder squirrel.SelectBuilder, totalCount ...*int64) (*[]{{.upperStartCamelObject}}, error) {
	var resp []{{.upperStartCamelObject}}

	query, _, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}
	
	if err = m.conn.QueryRowCtx(ctx, &resp, query);err != nil{
		return nil, err
	}

	if len(totalCount) != 0 {
		count := struct{Count int64 {{.countTag}}}{}
		query, _, err =sqlBuilder.Delete(selectBuilder, "Columns").(squirrel.SelectBuilder).Columns("COUNT(id) as count").ToSql()
		if err = m.conn.QueryRowCtx(ctx, &count, query);err != nil {
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
