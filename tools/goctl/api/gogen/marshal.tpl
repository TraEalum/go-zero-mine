// ----------------create----------------
func (r *Create{{.upperStartCamelObject}}Resp) Marshal(p *proto.{{.upperStartCamelObject}}) error {
    r.Id = p.Id

	return nil
}


func (r *{{.upperStartCamelObject}}) Unmarshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.unmarshallFields}}

	return nil
}


// ----------------query----------------
func (r *Query{{.upperStartCamelObject}}Resp) Marshal(p *proto.{{.upperStartCamelObject}}List) error {
    r.CurrPage = p.CurPage
    r.TotalPage = p.TotalPage
    r.TotalCount = p.PageSize
	Marshal{{.upperStartCamelObject}}Lst(&r.Data,p.{{.upperStartCamelObject}})

	return nil
}


func (r *Query{{.upperStartCamelObject}}Req) (p *proto.{{.upperStartCamelObject}}Filter) error {
    r.Id = p.Id
    r.PageNo = p.PageNo
    r.PageSize = p.PageSize
	return nil
}


// ----------------{{.upperStartCamelObject}}----------------
func (r *{{.upperStartCamelObject}}) Marshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.marshalFields}}

	return nil
}



func Marshal{{.upperStartCamelObject}}Lst(r *[]{{.upperStartCamelObject}},p []*proto.{{.upperStartCamelObject}}){
	for _,item := range p {
		var tmp {{.upperStartCamelObject}}
		tmp.Marshal(c)
		*r = append(*r,tmp)
	}
}
// END