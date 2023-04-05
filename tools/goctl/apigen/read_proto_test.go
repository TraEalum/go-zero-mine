package apigen

import (
	"bufio"
	"fmt"
	"github.com/serenize/snaker"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/emicklei/proto"
	"github.com/zeromicro/go-zero/tools/goctl/rpc/parser"
)

func TestMain(m *testing.M) {
	filePath := "product.proto"

	r, err := os.Open(filePath)
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
			ret.Message = append(ret.Message, parser.Message{Message: m})
		}),
	)

	f, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	buf := bufio.NewReader(f)
	var strs []string

Loop:
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return
			}
		}

		if strings.Contains(line, "Api Struct Gen") {
			for {
				line, _ := buf.ReadString('\n')
				if strings.Contains(line, "Struct Gen End") {
					break Loop
				}

				str := strings.Replace(line, "//", "", -1)
				str = strings.TrimSpace(str)
				strs = append(strs, str)
			}
		}

	}

	var s Schema

	for _, v := range ret.Message {
		// 判断是否在生成的api gen struct里面
		if !isInSlice(strs, v.Message.Name) {
			continue
		}
		message := &Message{
			Name:   v.Message.Name,
			Fields: make([]MessageField, 0, len(v.Message.Elements)),
		}
		fmt.Println(v.Message.Name)
		for _, ele := range v.Message.Elements {
			field, ok := (ele).(*proto.NormalField)
			if !ok {
				continue
			}
			//  注释
			var comment string
			if field.InlineComment != nil {
				comment = strings.Join(field.InlineComment.Lines, comment)
			}

			//数组判断
			var typ = field.Type
			if field.Repeated {
				typ = fmt.Sprintf("[]*%s", field.Type)
			}

			// 首字母大写
			paraName := stringx.From(field.Name).FirstUpper()

			apiField := NewMessageField(typ, paraName, comment,
				snaker.CamelToSnake(field.Name))

			message.AppendField(apiField)
		}

		s.CusMessages = append(s.CusMessages, message)
	}
}
