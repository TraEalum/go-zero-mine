package template

const (
	//FindListByTrans find list method
	FindListByTrans = `
func (m *default{{.upperStartCamelObject}}Model) FindListByTrans(ctx context.Context,session sqlx.Session, selectBuilder squirrel.SelectBuilder)(*[]{{.upperStartCamelObject}}, error){
	var resp []{{.upperStartCamelObject}}
	if session == nil{
		return nil,fmt.Errorf("session can not nil")
	}
	query, values, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	if err = session.QueryRowsPartialCtx(ctx, &resp, query, values...); err != nil {
		return nil, err
	}

	return &resp,nil
}
	`
	//FindListByTransMethod defines find a row method
	FindListByTransMethod = `FindListByTrans(ctx context.Context,session sqlx.Session, selectBuilder squirrel.SelectBuilder)(*[]{{.upperStartCamelObject}}, error)`
)
