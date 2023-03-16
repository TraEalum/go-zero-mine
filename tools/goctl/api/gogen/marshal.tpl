package types

{{.importProto}}

// Generated Start. Don't edit in this field.
func (r *{{.upperStartCamelObject}}) Marshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.marshalFields}}

	return nil
}

func (r *{{.upperStartCamelObject}}) Unmarshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.unmarshallFields}}

	return nil
}

func (r *Update{{.upperStartCamelObject}}Req) Unmarshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.unmarshallFields}}

	return nil
}

func (r *Query{{.upperStartCamelObject}}Resp) Marshal(p *proto.{{.upperStartCamelObject}}List) error {
    r.CurrPage = p.CurPage
    r.TotalPage = p.TotalPage
    r.TotalCount = p.TotalCount

	Marshal{{.upperStartCamelObject}}Lst(&r.{{.upperStartCamelObject}}List,p.{{.upperStartCamelObject}})

	return nil
}

func (r *Query{{.upperStartCamelObject}}Req) Unmarshal(p *proto.{{.upperStartCamelObject}}Filter) error {
    p.PageNo = r.PageNo
    p.PageSize = r.PageSize

	return nil
}

func Marshal{{.upperStartCamelObject}}Lst(r *[]{{.upperStartCamelObject}},p []*proto.{{.upperStartCamelObject}}){
	*r=[]{{.upperStartCamelObject}}{}  // 2022-11-11 fix object init not nil

	for _,item := range p {
		var tmp {{.upperStartCamelObject}}
		tmp.Marshal(item)
		*r = append(*r,tmp)
	}
}
// Generated End. Please do not delete this line.