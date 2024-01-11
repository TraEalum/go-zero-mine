package generator

const commonHead = `
// {{.method}} {{if .hasComment}}{{.comment}}{{end}}
// Code generated by goctl. just once,if file don't exist
// tpl src:tools/goctl/rpc/generator/curdGen.go`

const CreateLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if in.Get{{.pK}}() != {{.pV}} {
		if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.Get{{.pK}}()); err != sqlc.ErrNotFound {
			logx.WithContext(l.ctx).Infof("%v is already exists")

			return &{{.responseType}}{ {{.pK}}: in.Get{{.pK}}()}, nil
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
	
	return &{{.responseType}}{ {{.pK}} :uint64(id) }, nil
}
`

const DeleteLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error
	id := in.Get{{.pK}}()

	// delete
	if err = l.svcCtx.{{.modelName}}Model.Delete(l.ctx, nil, &model.{{.modelName}}{ {{.pK}}: &id}); err != nil {
		return nil, errorm.New(errorm.RecordDeleteFailed, "delete data fail.%v", err)
	}

	return &{{.responseType}}{}, nil
}
`

const UpdateLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.Get{{.pK}}()); err != nil && err != sqlc.ErrNotFound {
		logx.WithContext(l.ctx).Infof("find data fail. %v", err)

		return nil, err
	}else if err == sqlc.ErrNotFound{
		err = errorm.New(errorm.RecordNotFound, "{{.pK}} %v dose not exists.", in.Get{{.pK}}())
		logx.WithContext(l.ctx).Infof("find data fail. %v", err)

		return nil, err
	}

	id := in.Get{{.pK}}()
	where := model.{{.modelName}}{
		 {{.pK}}: &id,
	}
	{{.modelNameFirstLower}} := model.{{.modelName}}{}
	{{.modelNameFirstLower}}.Marshal(in)
	builder := util.NewUpdateBuiler(util.WithTable(where.TableName())).Where(&where).Updates(&{{.modelNameFirstLower}})

	// update
	if err = l.svcCtx.{{.modelName}}Model.Update(l.ctx, nil, builder.UpdateBuilder); err != nil {
		logx.WithContext(l.ctx).Infof("update fail. %v", err)

		return nil, errorm.New(errorm.RecordCreateFailed, "create data fail.%v", err)
	}

	return &{{.responseType}}{ {{.pK}}: in.Get{{.pK}}() }, nil
}
`

const QueryLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error
	var totalCount int64
	var {{.modelNameFirstLower}}List *[]model.{{.modelName}}
	resp := proto.{{.modelName}}List{
		{{.modelName}}: []*proto.{{.modelName}}{},
	}

	// build where
	id := in.Get{{.pK}}()
	where := model.{{.modelName}}{
		 {{.pK}}: &id,
	}
	builder := util.NewSelectBuilder(util.WithTable(where.TableName())).
		Where(&where).
		Limit(in)

	// query
	if {{.modelNameFirstLower}}List, err = l.svcCtx.{{.modelName}}Model.FindList(l.ctx, builder.SelectBuilder, &totalCount); err != nil {
		logx.WithContext(l.ctx).Infof("FindList fail. %v", err)
		
		return nil, errorm.New(errorm.RecordFindFailed, "FindList fail.%v", err)
	}

	model.Unmarshal{{.modelName}}Lst(&resp.{{.modelName}}, *{{.modelNameFirstLower}}List)

	// 分页
	resp.Total = totalCount
	resp.PerPage = int32(in.GetPageNo())
	resp.TotalPage = totalCount / in.GetPageSize()
	resp.Count = int32(len(resp.{{.modelName}}))
	resp.PerSize = int32(in.GetPageSize())

	if totalCount%in.GetPageSize() != 0 {
		resp.TotalPage += 1
	}

	return &resp, nil
}
`
const QueryDetailLogic = commonHead + `
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error
	var {{.modelNameFirstLower}} *model.{{.modelName}}
	resp := {{.responseType}}{}

	// query
	if {{.modelNameFirstLower}}, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.Get{{.pK}}()); err != nil {
		return nil, errorm.New(errorm.RecordFindFailed, "FindOne fail.%v", err)
	}

	{{.modelNameFirstLower}}.Unmarshal(&resp)

	return &resp, nil
}
`
