package parser

import (
	"fmt"
	goformat "go/format"
	"os"
	"strings"
	"testing"

	"github.com/emicklei/proto"
)

type Object struct {
	Api       string
	Type      string
	Marshal   string
	UnMarshal string
}

func TestParseMessage(t *testing.T) {
	path := "message.txt"
	b, err := os.OpenFile(path, os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}
	defer b.Close()

	p := proto.NewParser(b)
	set, err := p.Parse()
	if err != nil {
		panic(err)
	}

	var ret Proto
	proto.Walk(
		set,
		proto.WithMessage(func(m *proto.Message) {
			ret.Message = append(ret.Message, Message{Message: m})
		}),
	)

	for _, v := range ret.Message {
		parse(v)
	}

}

func parse(msg Message) string {
	var buf strings.Builder

	buf.WriteString(getApi(msg))
	buf.WriteString(getType(msg))
	buf.WriteString(getMarshal(msg))
	buf.WriteString(getUmMarshal(msg))

	res := buf.String()
	return res
}

func getType(msg Message) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("type %s struct { \n", msg.Message.Name))

	for _, v := range msg.Message.Elements {
		element := v.(*proto.NormalField)
		file := element.Field
		arr := ""
		//判断是否为数组
		if element.Repeated {
			arr = "[]"
			if !isSimpleType(element.Type) {
				arr = "[] *"
			}
		}

		buf.WriteString(fmt.Sprintf("%s  %s %s  \n", firstToUpper(file.Name), arr, file.Type))
	}

	buf.WriteString("}\n\n")

	return formatCode(buf.String())
}

func getApi(msg Message) string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("type %s  { \n", msg.Message.Name))

	for _, v := range msg.Message.Elements {
		element := v.(*proto.NormalField)
		file := element.Field
		arr := ""
		//判断是否为数组
		if element.Repeated {
			arr = "[]"
		}

		buf.WriteString(fmt.Sprintf("\t%s  %s%s  `json:\"%s\"` \n", firstToUpper(file.Name), arr, file.Type, file.Name))
	}

	buf.WriteString("}\n\n")

	return formatCode(buf.String())
}

func getMarshal(msg Message) string {
	var buf strings.Builder
	name := firstToUpper(msg.Name)

	buf.WriteString(fmt.Sprintf("func (m *%s) Marshal(p *%s) {\n", name, name))

	for _, v := range msg.Message.Elements {
		element := v.(*proto.NormalField)
		file := element.Field
		marshalName := firstToUpper(file.Name)
		buf.WriteString(fmt.Sprintf("m.%s = p.%s \n", marshalName, marshalName))
	}
	buf.WriteString("}\n\n")

	return formatCode(buf.String())
}

func getUmMarshal(msg Message) string {
	var buf strings.Builder

	name := firstToUpper(msg.Name)

	buf.WriteString(fmt.Sprintf("func (m *%s) Unmarshal(p *%s) {\n", name, name))

	for _, v := range msg.Message.Elements {
		element := v.(*proto.NormalField)
		file := element.Field
		marshalName := firstToUpper(file.Name)
		buf.WriteString(fmt.Sprintf("p.%s = m.%s\n", marshalName, marshalName))
	}
	buf.WriteString("}\n\n")

	return formatCode(buf.String())
}

func firstToUpper(res string) string {
	if res == "" {
		return ""
	}

	return strings.ToUpper(res[:1]) + res[1:]
}

func formatCode(code string) string {
	ret, err := goformat.Source([]byte(code))
	if err != nil {
		return code
	}

	return string(ret)
}

func isSimpleType(tp string) bool {
	arr := []string{"uint8", "uint16", "uint32", "uint64", "int", "int16", "int32", "int64", "string", "float32", "float64", "byte"}

	for _, v := range arr {
		if strings.ToLower(tp) == v {
			return true
		}
	}

	return false
}
