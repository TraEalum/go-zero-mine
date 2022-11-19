package types

{{.importProto}}

import (
	"fmt"
	"go-service/comm/define"
	"reflect"
)

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
    var (
    		v = reflect.ValueOf(r).Elem()
    		t = reflect.TypeOf(r).Elem()
    	)

    	for i := 0; i < v.NumField(); i++ {

    		tt := t.Field(i).Type.Elem()

    		if fmt.Sprintf("%v", v.Field(i).Interface()) == "<nil>" {

    			switch tt.Kind() {
    			case reflect.Int64:
    				var mark int64 = define.UpdateZeroMark
    				// 为空指针赋值
    				v.Field(i).Set(reflect.ValueOf(&mark))

    			case reflect.String:
    				var mark string = define.UpdateNullMark
    				v.Field(i).Set(reflect.ValueOf(&mark))

    			}
    		}
    	}


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
	*r=[]{{.upperStartCamelObject}}{}  // 2022-11-11 fix object init not nil
	for _,item := range p {
		var tmp {{.upperStartCamelObject}}
		tmp.Marshal(item)
		*r = append(*r,tmp)
	}
}

//TheEndLine   please do not delete this line