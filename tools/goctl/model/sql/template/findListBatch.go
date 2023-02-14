package template

const (
	FindListBatch = `
	//id不连续的时候有bug，慎用
func (m *default{{.upperStartCamelObject}}Model)FindListBatch(ctx context.Context,selectBuilder squirrel.SelectBuilder)(*[]{{.upperStartCamelObject}}, error){
	var resp []{{.upperStartCamelObject}}
	var maxId *int64
	var minId *int64
	var limit int64

	_, _, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	count := struct {
		MaxId int64 {{.maxTag}}
		MinId int64 {{.minTag}}
		}{}

	query, values, err := selectBuilder.Columns("MAX(id) as MaxId").Column("MIN(id) as MinId").ToSql()
	if err = m.conn.QueryRowCtx(ctx, &count, query, values...); err != nil {
		return nil, err
	}
	maxId = &count.MaxId
	minId = &count.MinId

	var batchSize int64 = 1000

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
		limit = ct
	}

	//batch search
	for *minId < *maxId {
		var temp []{{.upperStartCamelObject}}
		end :=*minId+batchSize
		if end > maxId{
			end = maxId
		}
		
		query, values, _ = selectBuilder.Where("id between ? and ?",minId, end).ToSql()

		err = m.conn.QueryRowsCtx(ctx, &temp, query, values...)
		if err != nil && err != ErrNotFound{
			return nil,err
		}
		if err == ErrNotFound{
			return &resp,nil
		}

		*minId = *minId+batchSize
		resp = append(resp, temp...)

		//if origin sql had limit condition
		if b && len(resp)>=int(limitSize){
			return &resp,nil
		}
	}


	return &resp,nil
}
	`
	FindListBatchMethod = `FindListBatch(ctx context.Context,selectBuilder squirrel.SelectBuilder)(*[]{{.upperStartCamelObject}}, error)`
)
