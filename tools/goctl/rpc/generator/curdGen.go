package generator

const commonHead = `
// {{.method}} {{if .hasComment}}{{.comment}}{{end}}
// Code generated by goctl. just once,if file don't exist
// tpl src:tools/goctl/rpc/generator/curdGen.go`

const CreateLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if in.{{.pK}} != {{.pV}} {
		if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.{{.pK}}); err != sqlc.ErrNotFound {
			logx.WithContext(l.ctx).Infof("%v is already exists")
			return &{{.responseType}}{ {{.pK}}: in.{{.pK}}}, nil
		}
	}

	// create
	{{.modelNameFirstLower}} := model.{{.modelName}}{}
	{{.modelNameFirstLower}}.Marshal(in)
	res,err:=l.svcCtx.{{.modelName}}Model.Insert(l.ctx, nil, &{{.modelNameFirstLower}})
	if  err != nil {
		return nil, errorm.New(errorm.RecordCreateFailed, "create data fail.%v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		logx.Error(err)
	}
	return &{{.responseType}}{ {{.pK}} :id }, nil
}
`

const DeleteLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// delete
	if err = l.svcCtx.{{.modelName}}Model.Delete(l.ctx, nil, &model.{{.modelName}}{ {{.pK}}: in.{{.pK}}}); err != nil {
		return nil, errorm.New(errorm.RecordDeleteFailed, "delete data fail.%v", err)
	}

	return &{{.responseType}}{}, nil
}
`

const UpdateLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.{{.pK}}); err != nil && err != sqlc.ErrNotFound {
		logx.WithContext(l.ctx).Infof("find data fail. %v", err)
		return nil, err
	}else if err == sqlc.ErrNotFound{
		err = errorm.New(errorm.RecordNotFound, "{{.pK}} %v dose not exists.", in.{{.pK}})
		logx.WithContext(l.ctx).Infof("find data fail. %v", err)
		return nil, err
	}

	where := model.{{.modelName}}{
		 {{.pK}}: in.{{.pK}},
	}
	{{.modelNameFirstLower}} := model.{{.modelName}}{}
	{{.modelNameFirstLower}}.Marshal(in)
	builder := util.NewUpdateBuiler(util.WithTable(where.TableName())).Where(&where).Updates(&{{.modelNameFirstLower}})

	// update
	if err = l.svcCtx.{{.modelName}}Model.Update(l.ctx, nil, builder.UpdateBuilder); err != nil {
		logx.WithContext(l.ctx).Infof("update fail. %v", err)
		return nil, errorm.New(errorm.RecordCreateFailed, "create data fail.%v", err)
	}

	return &{{.responseType}}{ {{.pK}}: in.{{.pK}} }, nil
}
`

const QueryLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error
	var totalCount int64
	resp := proto.{{.modelName}}List{
		{{.modelName}}: []*proto.{{.modelName}}{},
	}

	// build where
	where := model.{{.modelName}}{
		 {{.pK}}: in.{{.pK}},
	}
	builder := util.NewSelectBuilder(util.WithTable(where.TableName())).
		Where(&where).
		Limit(in)

	// query
	{{.modelNameFirstLower}}List := &[]model.{{.modelName}}{}
	if {{.modelNameFirstLower}}List, err = l.svcCtx.{{.modelName}}Model.FindList(l.ctx, builder.SelectBuilder, &totalCount); err != nil {
		logx.WithContext(l.ctx).Infof("FindList fail. %v", err)
		return nil, errorm.New(errorm.RecordFindFailed, "FindList fail.%v", err)
	}

	model.Unmarshal{{.modelName}}Lst(&resp.{{.modelName}}, *{{.modelNameFirstLower}}List)

	// 分页
	resp.TotalCount = totalCount
	resp.CurPage = in.GetPageNo()
	resp.TotalPage = totalCount / in.GetPageSize()
	if totalCount%in.GetPageSize() != 0 {
		resp.TotalPage += 1
	}

	return &resp, nil
}
`
const QueryDetailLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error
	resp := {{.responseType}}{}

	// query
	{{.modelNameFirstLower}} := &model.{{.modelName}}{}
	if {{.modelNameFirstLower}}, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.{{.pK}}); err != nil {
		return nil, errorm.New(errorm.RecordFindFailed, "FindOne fail.%v", err)
	}

	{{.modelNameFirstLower}}.Unmarshal(&resp)

	return &resp, nil
}
`
