package template

const (
	FindListBatch = `
func (m *default{{.upperStartCamelObject}}Model)FindListBatch(ctx context.Context,selectBuilder squirrel.SelectBuilder,sortById bool)(*[]{{.upperStartCamelObject}}, error){
	_, _, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	if sortById {
		return findListSortById(ctx,selectBuilder,m.conn)
	}

	return findListBatch(ctx,selectBuilder,m.conn)
}

func findListSortById(ctx context.Context,selectBuilder squirrel.SelectBuilder,conn sqlx.SqlConn) (*[]{{.upperStartCamelObject}},error){
	var resp []{{.upperStartCamelObject}}
	var maxId *int64
	var minId *int64
	var limitSize int64

	count := struct {
		MaxId int64 {{.maxTag}}
		MinId int64 {{.minTag}}
		}{}

	query, values, err := selectBuilder.Columns("MAX(id) as MaxId").Column("MIN(id) as MinId").ToSql()
	if err = conn.QueryRowCtx(ctx, &count, query, values...); err != nil {
		return nil, err
	}
	maxId = &count.MaxId
	minId = &count.MinId

	var batchSize int64 = 1000

	//if origin sql have limit
	limit, b := sqlBuilder.Get(selectBuilder, "Limit")
	if b {
		c, _ := limit.(string)
		ct, _ := strconv.ParseInt(c, 10, 64)
		limitSize = ct
	}

	//batch search
	for *minId < *maxId {
		var temp []{{.upperStartCamelObject}}
		end :=*minId+batchSize
		if end > *maxId{
			end = *maxId
		}
		
		query, values, _ = selectBuilder.Where("id between ? and ?",minId, end).ToSql()

		err = conn.QueryRowsCtx(ctx, &temp, query, values...)
		if err != nil && err != ErrNotFound {
			return nil,err
		}
		if err == ErrNotFound{
			return &resp,nil
		}

		*minId = *minId+batchSize
		resp = append(resp, temp...)

		//if origin sql had limit condition
		if b && len(resp)>=int(limitSize) {
			resp = resp[:limitSize]
			return &resp,nil
		}
	}

	return &resp,nil
}

func findListBatch(ctx context.Context,selectBuilder squirrel.SelectBuilder,conn sqlx.SqlConn) (*[]{{.upperStartCamelObject}},error){
	var resp []{{.upperStartCamelObject}}
	var totalCount *int64

	_, _, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	count := struct{Count int64 {{.countTag}}}{}

	query, values, err := sqlBuilder.Delete(selectBuilder, "Columns").(squirrel.SelectBuilder).Columns("COUNT(*) as count").ToSql()
	if err = conn.QueryRowCtx(ctx, &count, query, values...); err != nil {
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


	//batch search
	for startIndex <= *totalCount {
		var temp []{{.upperStartCamelObject}}
		
		query, values, _ = selectBuilder.Offset(uint64(startIndex)).Limit(uint64(batchSize)).ToSql()

		err = conn.QueryRowsCtx(ctx, &temp, query, values...)
		if err != nil && err != ErrNotFound{
			return nil,err
		}
		if err == ErrNotFound{
			return &resp,nil
		}

		resp = append(resp, temp...)
		if len(resp) >= int(*totalCount) {
			length :=*totalCount
			resp = resp[:length]
			return &resp,nil
		}

		startIndex = startIndex+batchSize
	}


	return &resp,nil
}
	`
	FindListBatchMethod = `FindListBatch(ctx context.Context,selectBuilder squirrel.SelectBuilder,sortById bool)(*[]{{.upperStartCamelObject}}, error)`
)
