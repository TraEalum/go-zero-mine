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
    r.Pagination = Pagination{
    		PerPage:   p.PerPage,
    		TotalPage: p.TotalPage,
    		Total:     p.Total,
    		PerSize:   p.PerSize,
    		Count:     p.Count,
    	}

	Marshal{{.upperStartCamelObject}}Lst(&r.{{.upperStartCamelObject}}List,p.{{.upperStartCamelObject}})

	return nil
}

func (r *Query{{.upperStartCamelObject}}Req) Unmarshal(p *proto.{{.upperStartCamelObject}}Filter) error {
    p.PageNo = r.PageNo
    p.PageSize = r.PageSize

	return nil
}

func Marshal{{.upperStartCamelObject}}Lst(r *[]{{.upperStartCamelObject}},p []*proto.{{.upperStartCamelObject}}){
	*r = make([]{{.upperStartCamelObject}}, 0, len(p))

	for _,item := range p {
		var tmp {{.upperStartCamelObject}}
		tmp.Marshal(item)
		*r = append(*r,tmp)
	}
}
// Generated End. Please do not delete this line.