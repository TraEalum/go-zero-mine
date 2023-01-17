package template

const (
	FindListBatch = `
func (m *default{{.upperStartCamelObject}}Model)FindListBatch(ctx context.Context,selectBuilder squirrel.SelectBuilder)(*[]{{.upperStartCamelObject}}, error){
	var resp []{{.upperStartCamelObject}}
	var totalCount *int64

	_, _, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	count := struct{Count int64 {{.countTag}}}{}

	query, values, err := sqlBuilder.Delete(selectBuilder, "Columns").(squirrel.SelectBuilder).Columns("COUNT(*) as count").ToSql()
	if err = m.conn.QueryRowCtx(ctx, &count, query, values...); err != nil {
		return nil, err
	}
	totalCount = &count.Count

	var batchSize int64 = 1000
	var startIndex int64 = 0

	//if origin sql have limit offset
	offset, b := sqlBuilder.Get(selectBuilder, "Offset")
	if b {
		index, _ := offset.(string)
		idx, _ := strconv.ParseInt(index, 10, 64)
		startIndex = idx
	}

	limit, b := sqlBuilder.Get(selectBuilder, "Limit")
	if b {
		c, _ := limit.(string)
		ct, _ := strconv.ParseInt(c, 10, 64)
		*totalCount = ct
	}

	//batch search
	for startIndex <= *totalCount{
		var temp []{{.upperStartCamelObject}}
		limitSize := startIndex+batchSize
		if limitSize > *totalCount {
			limitSize = *totalCount
		}
		query, values, _ = selectBuilder.Offset(uint64(startIndex)).Limit(uint64(limitSize)).ToSql()

		err = m.conn.QueryRowsCtx(ctx, &temp, query, values...)
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
