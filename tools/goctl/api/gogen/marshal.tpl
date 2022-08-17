package types

{{.importProto}}

//start
// ----------------create----------------
func (r *Create{{.upperStartCamelObject}}Resp) Marshal(p *proto.{{.upperStartCamelObject}}) error {
    r.Id = p.Id

	return nil
}


func (r *{{.upperStartCamelObject}}) Unmarshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.unmarshallFields}}

	return nil
}

// ----------------update----------------
func (r *Update{{.upperStartCamelObject}}Resp) Marshal(p *proto.{{.upperStartCamelObject}}) error {
	r.Id = p.Id

	return nil
}


func (r *Update{{.upperStartCamelObject}}Req) Unmarshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.unmarshallFields}}

	return nil
}

// ----------------query----------------
func (r *Query{{.upperStartCamelObject}}Resp) Marshal(p *proto.{{.upperStartCamelObject}}List) error {
    r.CurrPage = p.CurPage
    r.TotalPage = p.TotalPage
    r.TotalCount = p.TotalCount
	Marshal{{.upperStartCamelObject}}Lst(&r.{{.upperStartCamelObject}}List,p.{{.upperStartCamelObject}})

	return nil
}


func (r *Query{{.upperStartCamelObject}}Req) Unmarshal(p *proto.{{.upperStartCamelObject}}Filter) error {
    p.Id = r.Id
    p.PageNo = r.PageNo
    p.PageSize = r.PageSize
	return nil
}


// ----------------marshal----------------
func (r *{{.upperStartCamelObject}}) Marshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.marshalFields}}

	return nil
}



func Marshal{{.upperStartCamelObject}}Lst(r *[]{{.upperStartCamelObject}},p []*proto.{{.upperStartCamelObject}}){
	for _,item := range p {
		var tmp {{.upperStartCamelObject}}
		tmp.Marshal(item)
		*r = append(*r,tmp)
	}
}

//TheEndLine   please do not delete this line