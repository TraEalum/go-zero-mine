package apigen

import (
	"fmt"
	"os"
	"testing"

	"github.com/emicklei/proto"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
)


func TestMain(m *testing.T){
	r,err := os.Open("product.proto")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer r.Close()

	p := proto.NewParser(r)

	set, err := p.Parse()
	if err != nil {
		fmt.Println(err)
	}

	var ret parser.Proto

	proto.Walk(set,
	proto.WithMessage(func(m *proto.Message) {
		ret.Message = append(ret.Message,parser.Message{Message: m})
	}),
	)

	for _, v := range ret.Message{
		temp := v
		fmt.Println(temp.Message.Name)
		fmt.Println(temp.Message.Elements)

		ele := temp.Message.Elements

		for _, e := range ele {
			fmt.Println(e.)
		}
	}
}