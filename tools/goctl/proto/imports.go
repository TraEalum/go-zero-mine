package proto

import (
	"bytes"
	"fmt"
)

func (s *Schema) setPackage(buf *bytes.Buffer) {
	buf.WriteString(fmt.Sprintf("syntax = \"%s\";\n", s.Syntax))
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("option go_package =\"%s\";\n", s.GoPackage))
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("package %s;\n", s.Package))
	buf.WriteString("\n")
}
