package template

const (
	FindListBatch = `
func (m *default{{.upperStartCamelObject}}Model)FindListBatch(ctx context.Context,selectBuilder squirrel.SelectBuilder)(*[]{{.upperStartCamelObject}}, error){
	var resp []{{.upperStartCamelObject}}
	var totalCount *int64

	selectBuilder.RemoveLimit()
	query, values, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	count := struct{Count int64 {{.countTag}}}{}

	query, values, err = sqlBuilder.Delete(selectBuilder, "Columns").(squirrel.SelectBuilder).Columns("COUNT(*) as count").ToSql()
	if err = m.conn.QueryRowCtx(ctx, &count, query, values...); err != nil {
		return nil, err
	}
	totalCount = &count.Count

	var batchSize int64 = 1000
	var startIndex int64 = 0

	//if origin sql have limit offset
	offset, b := sqlBuilder.Get(selectBuilder, "Offset")
	if b {
		startIndex = offset.(int64)
	}

	limit, b := sqlBuilder.Get(selectBuilder, "Limit")
	if b {
		*totalCount = limit.(int64)
	}

	//batch search
	for startIndex <= *totalCount{
		var temp []{{.upperStartCamelObject}}
		limitSize := startIndex+batchSize
		if limitSize > *totalCount {
			limitSize = *totalCount
		}
		query, values, err = selectBuilder.Offset(uint64(startIndex)).Limit(uint64(limitSize)).ToSql()
		if  err != nil {
			return nil,err
		}

		err = m.conn.QueryRowCtx(ctx, &temp, query, values...)
		if err != nil && err != ErrNotFound{
			return nil,err
		}
		if err == ErrNotFound{
			return &resp,nil
		}

		resp = append(resp, temp...)
		startIndex = startIndex+batchSize
	}


	return &resp,nil
}
	`
	FindListBatchMethod = `FindListBatch(ctx context.Context,selectBuilder squirrel.SelectBuilder)(*[]{{.upperStartCamelObject}}, error)`
)
