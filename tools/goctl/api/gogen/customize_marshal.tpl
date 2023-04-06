package types

{{.importProto}}


func (r *{{.upperStartCamelObject}}) Unmarshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.unmarshallFields}}

	return nil
}


func (r *{{.upperStartCamelObject}}) Marshal(p *proto.{{.upperStartCamelObject}}) error {
	{{.marshalFields}}

	return nil
}



func Marshal{{.upperStartCamelObject}}Lst(r *[]{{.upperStartCamelObject}},p []*proto.{{.upperStartCamelObject}}){
	if r == nil {
	    return
	}

	for _,item := range p {
		var tmp {{.upperStartCamelObject}}
		tmp.Marshal(item)
		*r = append(*r,tmp)
	}
}

func UnMarshal{{.upperStartCamelObject}}Lst(r []{{.upperStartCamelObject}}, p *[]*proto.{{.upperStartCamelObject}}){
    if p == nil {
        return
    }

    for _, item := range r {
        var tmp proto.{{.upperStartCamelObject}}
        item.Unmarshal(&tmp)
        *p = append(*p,&tmp)
    }
}



// Generated End. Please do not delete this line.