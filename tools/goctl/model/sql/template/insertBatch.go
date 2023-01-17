package template

const (
	InsertBatch = `
func (m *default{{.upperStartCamelObject}}Model)InsertBatch(ctx context.Context,session sqlx.Session,dataList *[]{{.upperStartCamelObject}})(sql.Result,error){
	if dataList == nil && len(*dataList) == 0 {
		return nil, fmt.Errorf( "batch insert fail, dataList not set.")
	}

	query := fmt.Sprintf("insert into %s (%s) values", m.table, {{.lowerStartCamelObject}}RowsExpectAutoSet)
	var args  []interface{}
	var values []string

	for _, data := range *dataList {
		value := fmt.Sprintf("({{.expression}})")
		values = append(values, value)

		args = append(args, {{.expressionValues}})
	}

	if session != nil {
	 return session.ExecCtx(ctx, query+strings.Join(values, ","), args...)
	}

	return m.conn.ExecCtx(ctx, query+strings.Join(values, ","), args...)
}
`
	InsertBatchMethod = `InsertBatch(ctx context.Context,session sqlx.Session,dataList *[]{{.upperStartCamelObject}})(sql.Result,error)`
)
