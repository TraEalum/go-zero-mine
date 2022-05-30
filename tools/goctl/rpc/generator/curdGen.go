package generator

const CreateLogic = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if in.Id != 0 {
		if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.Id); err != sqlc.ErrNotFound {
			logx.WithContext(l.ctx).Infof("%v is already exists")
			return &{{.responseType}}{Id: in.Id}, nil
		}
	}

	// create
	{{.modelNameFirstLower}} := model.{{.modelName}}{}
	{{.modelNameFirstLower}}.Marshal(in)
	if _, err = l.svcCtx.{{.modelName}}Model.Insert(l.ctx, nil, &{{.modelNameFirstLower}}); err != nil {
		return nil, errorm.New(errorm.RecordCreateFailed, "create data fail.%v", err)
	}

	return &{{.responseType}}{}, nil
}
`

const DeleteLogic = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if in.Id != 0 {
		if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.Id); err != sqlc.ErrNotFound {
			logx.WithContext(l.ctx).Infof("%v is already exists")
			return &{{.responseType}}{Id: in.Id}, nil
		}
	}

	// create
	{{.modelNameFirstLower}} := model.{{.modelName}}{}
	{{.modelNameFirstLower}}.Marshal(in)
	if _, err = l.svcCtx.{{.modelName}}Model.Insert(l.ctx, nil, &{{.modelNameFirstLower}}); err != nil {
		return nil, errorm.New(errorm.RecordCreateFailed, "create data fail.%v", err)
	}

	return &{{.responseType}}{}, nil
}
`

const UpdateLogic = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if in.Id != 0 {
		if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.Id); err != sqlc.ErrNotFound {
			logx.WithContext(l.ctx).Infof("%v is already exists")
			return &{{.responseType}}{Id: in.Id}, nil
		}
	}

	// create
	{{.modelNameFirstLower}} := model.{{.modelName}}{}
	{{.modelNameFirstLower}}.Marshal(in)
	if _, err = l.svcCtx.{{.modelName}}Model.Insert(l.ctx, nil, &{{.modelNameFirstLower}}); err != nil {
		return nil, errorm.New(errorm.RecordCreateFailed, "create data fail.%v", err)
	}

	return &{{.responseType}}{}, nil
}
`
const QueryLogic = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if in.Id != 0 {
		if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.Id); err != sqlc.ErrNotFound {
			logx.WithContext(l.ctx).Infof("%v is already exists")
			return &{{.responseType}}{Id: in.Id}, nil
		}
	}

	// create
	{{.modelNameFirstLower}} := model.{{.modelName}}{}
	{{.modelNameFirstLower}}.Marshal(in)
	if _, err = l.svcCtx.{{.modelName}}Model.Insert(l.ctx, nil, &{{.modelNameFirstLower}}); err != nil {
		return nil, errorm.New(errorm.RecordCreateFailed, "create data fail.%v", err)
	}

	return &{{.responseType}}{}, nil
}
`
const QueryDetailLogic = `{{if .hasComment}}{{.comment}}{{end}}
func (l *{{.logicName}}) {{.method}} (in {{.request}}) ({{.response}}, error) {
	var err error

	// check whether it already exists
	if in.Id != 0 {
		if _, err = l.svcCtx.{{.modelName}}Model.FindOne(l.ctx, in.Id); err != sqlc.ErrNotFound {
			logx.WithContext(l.ctx).Infof("%v is already exists")
			return &{{.responseType}}{Id: in.Id}, nil
		}
	}

	// create
	{{.modelNameFirstLower}} := model.{{.modelName}}{}
	{{.modelNameFirstLower}}.Marshal(in)
	if _, err = l.svcCtx.{{.modelName}}Model.Insert(l.ctx, nil, &{{.modelNameFirstLower}}); err != nil {
		return nil, errorm.New(errorm.RecordCreateFailed, "create data fail.%v", err)
	}

	return &{{.responseType}}{}, nil
}
`
