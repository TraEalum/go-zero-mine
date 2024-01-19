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
	
	if err = m.conn.QueryRowsPartialCtx(ctx, &resp, query, values...);err != nil{
		return nil, err
	}

	if len(totalCount) == 0 || (len(totalCount) == 1 && totalCount[0] == nil) {
		totalCount[0] = new(int64)
		return &resp, nil
	}

	count := struct{Count int64 {{.countTag}}}{}
	
	//  type assertion must be checked
	sb, ok := sqlBuilder.Delete(selectBuilder, "Columns").(squirrel.SelectBuilder)
	if !ok {
		return nil, errNotSelectBuilder
	}

	query, values, err =sb.Columns("COUNT(*) as count").RemoveOffset().ToSql()
	if err != nil {
		return nil, err
	}

	if err = m.conn.QueryRowCtx(ctx, &count, query, values...);err != nil {
		return nil, err
	}

	*totalCount[0] = count.Count

	return &resp, nil
}
`

	// ListsMethod defines an interface method template for lists code in model
	ListsMethod = `FindList(ctx context.Context, selectBuilder squirrel.SelectBuilder, totalCount ...*int64) (*[]{{.upperStartCamelObject}},error)`
)
