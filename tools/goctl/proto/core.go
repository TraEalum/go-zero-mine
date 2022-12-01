package proto

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/tools/goctl/util"

	"github.com/chuckpreslar/inflect"
	"github.com/serenize/snaker"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

const (
	// proto3 is a describing the proto3 syntax type.
	proto3 = "proto3"

	// indent represents the indentation amount for fields. the style guide suggests
	// two spaces
	indent = "  "
)

// GenerateSchema generates a protobuf schema from a database connection and a package name.
// A list of tables to ignore may also be supplied.
// The returned schema implements the `fmt.Stringer` interface, in order to generate a string
// representation of a protobuf schema.
// Do not rely on the structure of the Generated schema to provide any context about
// the protobuf types. The schema reflects the layout of a protobuf file and should be used
// to pipe the output of the `Schema.String()` to a file.
func GenerateSchema(db *sql.DB, table string, ignoreTables []string, serviceName, goPkg, pkg string, dir string, subTableNumber int, subTableKey string) (*Schema, error) {
	s := &Schema{
		Dir: dir,
	}

	dbs, err := dbSchema(db)
	if nil != err {
		return nil, err
	}

	s.Syntax = proto3
	s.ServiceName = serviceName
	if "" != pkg {
		s.Package = pkg
	}
	if "" != goPkg {
		s.GoPackage = goPkg
	} else {
		s.GoPackage = "./" + s.Package
	}

	tmpName := stringx.From(table).ToCamelWithStartLower()
	humpTableName := strings.ToUpper(string(tmpName[0])) + tmpName[1:]

	s.HumpTbName = humpTableName

	cols, err := dbColumns(db, dbs, table, subTableNumber, subTableKey)
	if nil != err {
		return nil, err
	}

	// fmt.Println("log print:" + table)

	err = typesFromColumns(s, cols, ignoreTables)
	if nil != err {
		return nil, err
	}

	sort.Sort(s.Imports)
	sort.Sort(s.Messages)
	sort.Sort(s.Enums)

	return s, nil
}

// typesFromColumns creates the appropriate schema properties from a collection of column types.
func typesFromColumns(s *Schema, cols []Column, ignoreTables []string) error {
	messageMap := map[string]*Message{}
	ignoreMap := map[string]bool{}
	for _, ig := range ignoreTables {
		ignoreMap[ig] = true
	}

	for _, c := range cols {
		if _, ok := ignoreMap[c.TableName]; ok {
			continue
		}
		messageName := snaker.SnakeToCamel(c.TableName)
		// messageName = inflect.Singularize(messageName)

		// fmt.Printf("print table origin[%s] snaker[%s]", messageName, c.TableName)

		msg, ok := messageMap[messageName]
		if !ok {
			messageMap[messageName] = &Message{Name: messageName, Comment: c.TableComment}
			msg = messageMap[messageName]
		}

		err := parseColumn(s, msg, c)
		if nil != err {
			return err
		}
	}

	for _, v := range messageMap {
		s.Messages = append(s.Messages, v)
	}

	return nil
}

func dbSchema(db *sql.DB) (string, error) {
	var schema string

	err := db.QueryRow("SELECT SCHEMA()").Scan(&schema)

	return schema, err
}

func dbColumns(db *sql.DB, schema, table string, subTableNumber int, subTableKey string) ([]Column, error) {

	tableArr := strings.Split(table, ",")

	q := "SELECT c.TABLE_NAME, c.COLUMN_NAME, c.IS_NULLABLE, c.DATA_TYPE, " +
		"c.CHARACTER_MAXIMUM_LENGTH, c.NUMERIC_PRECISION, c.NUMERIC_SCALE, c.COLUMN_TYPE ,c.COLUMN_COMMENT,t.TABLE_COMMENT " +
		"FROM INFORMATION_SCHEMA.COLUMNS as c  LEFT JOIN  INFORMATION_SCHEMA.TABLES as t  on c.TABLE_NAME = t.TABLE_NAME and  c.TABLE_SCHEMA = t.TABLE_SCHEMA" +
		" WHERE c.TABLE_SCHEMA = ?"

	if table != "" && table != "*" {
		q += " AND c.TABLE_NAME IN('" + strings.TrimRight(strings.Join(tableArr, "' ,'"), ",") + "')"
	}

	q += " ORDER BY c.TABLE_NAME, c.ORDINAL_POSITION"

	rows, err := db.Query(q, schema)
	defer rows.Close()
	if nil != err {
		return nil, err
	}

	cols := []Column{}

	for rows.Next() {
		cs := Column{}
		err := rows.Scan(&cs.TableName, &cs.ColumnName, &cs.IsNullable, &cs.DataType,
			&cs.CharacterMaximumLength, &cs.NumericPrecision, &cs.NumericScale, &cs.ColumnType, &cs.ColumnComment, &cs.TableComment)
		if err != nil {
			log.Fatal(err)
		}

		if subTableNumber > 1 && subTableKey != "" {
			//从0下标开始，需要减一
			splitNum := util.GetSplitNum(subTableNumber - 1)
			//把最后得特殊符号"_"去掉，需要减一
			cs.TableName = cs.TableName[0 : len(cs.TableName)-splitNum-1]
		}

		if cs.TableComment == "" {
			cs.TableComment = stringx.From(cs.TableName).ToCamelWithStartLower()
		}

		cols = append(cols, cs)
	}
	if err := rows.Err(); nil != err {
		return nil, err
	}

	return cols, nil
}

// Schema is a representation of a protobuf schema.
type Schema struct {
	Syntax      string
	ServiceName string
	GoPackage   string
	Package     string
	Dir         string
	Imports     sort.StringSlice
	Messages    MessageCollection
	Enums       EnumCollection
	HumpTbName  string
}

// MessageCollection represents a sortable collection of messages.
type MessageCollection []*Message

func (mc MessageCollection) Len() int {
	return len(mc)
}

func (mc MessageCollection) Less(i, j int) bool {
	return mc[i].Name < mc[j].Name
}

func (mc MessageCollection) Swap(i, j int) {
	mc[i], mc[j] = mc[j], mc[i]
}

// EnumCollection represents a sortable collection of enums.
type EnumCollection []*Enum

func (ec EnumCollection) Len() int {
	return len(ec)
}

func (ec EnumCollection) Less(i, j int) bool {
	return ec[i].Name < ec[j].Name
}

func (ec EnumCollection) Swap(i, j int) {
	ec[i], ec[j] = ec[j], ec[i]
}

// AppendImport adds an import to a schema if the specific import does not already exist in the schema.
func (s *Schema) AppendImport(imports string) {
	shouldAdd := true
	for _, si := range s.Imports {
		if si == imports {
			shouldAdd = false
			break
		}
	}

	if shouldAdd {
		s.Imports = append(s.Imports, imports)
	}

}

func (s *Schema) String() string {
	log.Println("call .. func (s *Schema) String()", s.Dir)
	_, err := os.Stat(s.Dir)
	//如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
	if os.IsNotExist(err) {
		log.Println("call ..s.CreateString()")
		return s.CreateString()
	}
	log.Println("call ..s.UpdateString()")
	return s.UpdateString()
}

// String returns a string representation of a Schema.
func (s *Schema) CreateString() string {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("syntax = \"%s\";\n", s.Syntax))
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("option go_package =\"%s\";\n", s.GoPackage))
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("package %s;\n", s.Package))
	buf.WriteString("\n")
	buf.WriteString("// Already Exist Table:\n")
	for _, m := range s.Messages {
		buf.WriteString("// " + m.Name)
		buf.WriteString("\n")
	}
	buf.WriteString("// Exist Table End\n")
	buf.WriteString("\n")
	buf.WriteString("// Message Record Start\n")

	for _, m := range s.Messages {
		buf.WriteString("//--------------------------------" + m.Comment + "--------------------------------")
		buf.WriteString("\n")
		m.GenDefaultMessage(buf)
		m.GenRpcSearchReqMessage(buf, true)
	}
	buf.WriteString("// Message Record End\n")
	if len(s.Enums) > 0 {
		buf.WriteString("// Enums Record Start\n")
		for _, e := range s.Enums {
			buf.WriteString(fmt.Sprintf("%s\n", e))
		}
		buf.WriteString("// Enums Record End\n")
	}

	buf.WriteString("\n")
	buf.WriteString("// ------------------------------------ \n")
	buf.WriteString("// Rpc Func\n")
	buf.WriteString("// ------------------------------------ \n\n")

	funcTpl := "service " + s.ServiceName + "{\n"
	for _, m := range s.Messages {
		funcTpl += "\t //-----------------------" + m.Comment + "----------------------- \n"
		funcTpl += "\t rpc Create" + m.Name + "(" + m.Name + ") returns (" + m.Name + "); \n"
		funcTpl += "\t rpc Update" + m.Name + "(" + m.Name + ") returns (" + m.Name + "); \n"
		funcTpl += "\t rpc Delete" + m.Name + "(" + m.Name + ") returns (" + m.Name + "); \n"
		funcTpl += "\t rpc Query" + m.Name + "Detail(" + m.Name + "Filter) returns (" + m.Name + "); \n"
		funcTpl += "\t rpc Query" + m.Name + "List(" + m.Name + "Filter) returns (" + m.Name + "List); \n"
	}
	funcTpl = funcTpl + "\t // Service Record End\n"
	funcTpl = funcTpl + "}"
	buf.WriteString(funcTpl)
	err := ioutil.WriteFile(s.Dir, []byte(buf.String()), 0666)
	if err != nil {
		log.Println(err)
		return ""
	}
	return "DONE"
}

// String returns a string representation of a Schema.
func (s *Schema) UpdateString() string {
	bufNew := new(bytes.Buffer)
	file, err := os.OpenFile(s.Dir, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Sprintf("Open file error!%v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	var _ = stat.Size()
	endLine := ""
	buf := bufio.NewReader(file)

	//写已存在表名
	for {
		line, err := buf.ReadString('\n')
		bufNew.WriteString(line)
		if strings.Contains(line, "Already Exist Table") {
			break
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "Read file error!"
			}
		}
	}
	var existTableName []string

	for {
		line, err := buf.ReadString('\n')
		if strings.Contains(line, "Exist Table End") {
			endLine = line
			break
		}
		if len(line) < 3 {
			fmt.Println("could be error")
			continue
		}
		existTableName = append(existTableName, strings.TrimRight(line[3:], "\n"))
		bufNew.WriteString(line)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "Read file error!"
			}
		}
	}
	var newTableNames []string
	for _, m := range s.Messages {
		if !isInSlice(existTableName, m.Name) {
			newTableNames = append(newTableNames, m.Name)
			bufNew.WriteString("// " + m.Name + "\n")
		}
	}
	bufNew.WriteString(endLine)

	// 写Messages
	for {
		line, err := buf.ReadString('\n')
		if strings.Contains(line, "Message Record End") {
			endLine = line
			break
		}
		bufNew.WriteString(line)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "Read file error!"
			}
		}
	}

	for _, m := range s.Messages {
		//更新Default信息
		s.updateDefaultMessage(bufNew, m)

		//更新Filter信息
		s.updateFilterMessage(bufNew, m)

		if !isInSlice(existTableName, m.Name) {
			bufNew.WriteString("//--------------------------------" + m.Comment + "--------------------------------")
			bufNew.WriteString("\n")
			m.GenDefaultMessage(bufNew)
			m.GenRpcSearchReqMessage(bufNew, true)
		}
	}
	if len(s.Messages) > 0 {
		bufNew.WriteString(endLine)
	}

	// 写enum
	var existEnumText []string
	for {
		line, err := buf.ReadString('\n')
		if strings.Contains(line, "Enums Record End") {
			endLine = line
			break
		}

		if strings.Contains(line, "Rpc Func") {
			bufNew.WriteString(line)
			break
		}

		bufNew.WriteString(line)
		reg := regexp.MustCompile(`^enum [A-Za-z0-9]*`)
		enumText := reg.FindAllString(line, -1)
		if enumText != nil {
			existEnumText = append(existEnumText, enumText[0][5:])
		}
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "Read file error!"
			}
		}
	}

	if len(s.Enums) > 0 {
		for _, e := range s.Enums {
			if !isInSlice(existEnumText, e.Name) {
				bufNew.WriteString(fmt.Sprintf("%s\n", e))
			}
		}
		bufNew.WriteString(endLine)
	}

	// 写rpc服务名
	for {
		line, err := buf.ReadString('\n')
		if strings.Contains(line, "Service Record End") {
			break
		}
		bufNew.WriteString(line)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return "Read file error!"
			}
		}
	}

	funcTpl := ""
	for _, m := range s.Messages {
		if !isInSlice(existTableName, m.Name) {
			funcTpl += "\t //-----------------------" + m.Comment + "----------------------- \n"
			funcTpl += "\t rpc Create" + m.Name + "(" + m.Name + ") returns (" + m.Name + "); \n"
			funcTpl += "\t rpc Update" + m.Name + "(" + m.Name + ") returns (" + m.Name + "); \n"
			funcTpl += "\t rpc Delete" + m.Name + "(" + m.Name + ") returns (" + m.Name + "); \n"
			funcTpl += "\t rpc Query" + m.Name + "Detail(" + m.Name + "Filter) returns (" + m.Name + "); \n"
			funcTpl += "\t rpc Query" + m.Name + "List(" + m.Name + "Filter) returns (" + m.Name + "List); \n"
		}
	}
	funcTpl = funcTpl + "\t // Service Record End\n"
	funcTpl = funcTpl + "}"

	bufNew.WriteString(funcTpl)
	err = ioutil.WriteFile(s.Dir, []byte(bufNew.String()), 0666) //写入文件(字节数组)
	return "Done"
}

func (s *Schema) updateDefaultMessage(buf *bytes.Buffer, m *Message) {
	s.makeInstanceMessage(buf, m, "")
}

func (s *Schema) updateFilterMessage(buf *bytes.Buffer, m *Message) {
	s.makeInstanceMessage(buf, m, "Filter")
}

func (s *Schema) makeInstanceMessage(buf *bytes.Buffer, m *Message, extStr string) {
	oldTmp := buf.String()
	if oldTmp == "" || s.HumpTbName == "" {
		return
	}

	tmpName := stringx.From(s.HumpTbName).ToCamelWithStartLower()
	name := strings.ToUpper(string(tmpName[0])) + tmpName[1:]

	fmt.Println("name-----", name)
	var lastTag int

	//找到旧的内容,等下用来替换
	reg := fmt.Sprintf("message %s%s %s", name, extStr, "{[^}]+}")
	re := regexp.MustCompile(reg)
	oldSubStrings := re.FindStringSubmatch(oldTmp)

	if len(oldSubStrings) > 0 {
		tmpBuf := new(bytes.Buffer)

		switch extStr {
		case "Filter":
			m.GenRpcSearchReqMessage(tmpBuf, false)
		default:
			m.GenDefaultMessage(tmpBuf)
		}

		newTableString := strings.Replace(tmpBuf.String(), "}\n", "}", 1)

		tmpBuf.Reset()

		re = regexp.MustCompile("[1-9]\\d*")
		nums := re.FindAllString(newTableString, -1)

		if len(nums) > 0 {
			lastTag, _ = strconv.Atoi(nums[len(nums)-1])
		}

		//重新生成自定义标签
		var newCustomTag string
		reg = fmt.Sprintf("Custom Tag .You Can Edit.%s", "([^}]+})")
		re = regexp.MustCompile(reg)
		oldCustomStrings := re.FindStringSubmatch(oldSubStrings[0])
		if len(oldCustomStrings) > 0 && lastTag > 0 {
			oldEditTag := "Custom Tag .You Can Edit."
			newCustomTag = strings.Replace(newTableString, oldEditTag, oldEditTag+s.makeCustomStr(oldCustomStrings[1], lastTag), 1)
		} else {
			newCustomTag = newTableString
		}

		buf.Reset()
		newCustomTag = strings.Replace(newCustomTag, "}\n", "}", 1)
		newStr := strings.Replace(oldTmp, oldSubStrings[0], newCustomTag, 1)
		buf.WriteString(newStr)
	}

	return
}

func (s *Schema) makeCustomStr(oldCustomStr string, count int) string {
	if oldCustomStr == "" {
		return oldCustomStr
	}

	var newCustomStr string
	for _, v := range strings.Split(oldCustomStr, "\n") {
		reg := fmt.Sprintf("= (%s+)", "\\d")
		re := regexp.MustCompile(reg)

		tmp := re.FindStringSubmatch(v)
		if len(tmp) > 0 {
			count++

			oldStr := fmt.Sprintf("= %s", tmp[1])
			newStr := fmt.Sprintf("= %d", count)

			newCustomStr = fmt.Sprintf("%s\n%s", newCustomStr, strings.Replace(v, oldStr, newStr, 1))
		}

	}

	return newCustomStr
}

// Enum represents a protocol buffer enumerated type.
type Enum struct {
	Name    string
	Comment string
	Fields  []EnumField
}

// String returns a string representation of an Enum.
func (e *Enum) String() string {
	buf := new(bytes.Buffer)

	buf.WriteString(fmt.Sprintf("// %s \n", e.Comment))
	buf.WriteString(fmt.Sprintf("enum %s {\n", e.Name))

	for _, f := range e.Fields {
		buf.WriteString(fmt.Sprintf("%s%s;\n", indent, f))
	}

	buf.WriteString("}\n")

	return buf.String()
}

// AppendField appends an EnumField to an Enum.
func (e *Enum) AppendField(ef EnumField) error {
	for _, f := range e.Fields {
		if f.Tag() == ef.Tag() {
			return fmt.Errorf("tag `%d` is already in use by field `%s`", ef.Tag(), f.Name())
		}
	}

	e.Fields = append(e.Fields, ef)

	return nil
}

// EnumField represents a field in an enumerated type.
type EnumField struct {
	name string
	tag  int
}

// NewEnumField constructs an EnumField type.
func NewEnumField(name string, tag int) EnumField {
	name = strings.ToUpper(name)

	re := regexp.MustCompile(`([^\w]+)`)
	name = re.ReplaceAllString(name, "_")

	return EnumField{name, tag}
}

// String returns a string representation of an Enum.
func (ef EnumField) String() string {
	return fmt.Sprintf("%s = %d", ef.name, ef.tag)
}

// Name returns the name of the enum field.
func (ef EnumField) Name() string {
	return ef.name
}

// Tag returns the identifier tag of the enum field.
func (ef EnumField) Tag() int {
	return ef.tag
}

// newEnumFromStrings creates an enum from a name and a slice of strings that represent the names of each field.
func newEnumFromStrings(name, comment string, ss []string) (*Enum, error) {
	enum := &Enum{}
	enum.Name = name
	enum.Comment = comment

	for i, s := range ss {
		err := enum.AppendField(NewEnumField(s, i))
		if nil != err {
			return nil, err
		}
	}

	return enum, nil
}

// Service represents a protocol buffer service.
// TODO: Implement this in a schema.
type Service struct{}

// Message represents a protocol buffer message.
type Message struct {
	Name    string
	Comment string
	Fields  []MessageField
}

//gen default message
func (m Message) GenDefaultMessage(buf *bytes.Buffer) {
	mOrginName := m.Name
	mOrginFields := m.Fields

	curFields := []MessageField{}
	var filedTag int
	for _, field := range m.Fields {
		if isInSlice([]string{"version", "del_state", "delete_time"}, field.Name) {
			continue
		}
		filedTag++
		field.tag = filedTag
		field.Name = stringx.From(field.Name).ToCamelWithStartLower()
		if field.Comment == "" {
			field.Comment = field.Name
		}
		curFields = append(curFields, field)
	}
	m.Fields = curFields

	tmpStr := fmt.Sprintf("%s\n", m)

	//增加数据库字段开始结束标签, 自定义标签
	tmpStr = strings.Replace(tmpStr, "{", "{\n  //Database Tag Begin. DO NOT EDIT !!! ", 1)
	tmpStr = strings.Replace(tmpStr, "}", "  //Database Tag End. DO NOT EDIT!!!  \n\n  //Custom Tag .You Can Edit. \n\n}", 1)

	buf.WriteString(tmpStr)

	//reset
	m.Name = mOrginName
	m.Fields = mOrginFields
}

//gen add req message
func (m Message) GenRpcAddReqRespMessage(buf *bytes.Buffer) {
	mOrginName := m.Name
	mOrginFields := m.Fields

	//req
	m.Name = mOrginName
	curFields := []MessageField{}
	var filedTag int
	for _, field := range m.Fields {
		if isInSlice([]string{"id", "create_time", "update_time", "version", "del_state", "delete_time"}, field.Name) {
			continue
		}
		filedTag++
		field.tag = filedTag
		field.Name = stringx.From(field.Name).ToCamelWithStartLower()
		if field.Comment == "" {
			field.Comment = field.Name
		}
		curFields = append(curFields, field)
	}
	m.Fields = curFields
	buf.WriteString(fmt.Sprintf("%s\n", m))

	//reset
	m.Name = mOrginName
	m.Fields = mOrginFields

	//resp
	// m.Name = "Add" + mOrginName + "Resp"
	// m.Fields = []MessageField{}
	// buf.WriteString(fmt.Sprintf("%s\n", m))

	// //reset
	// m.Name = mOrginName
	// m.Fields = mOrginFields
}

//gen add resp message
func (m Message) GenRpcUpdateReqMessage(buf *bytes.Buffer) {
	mOrginName := m.Name
	mOrginFields := m.Fields

	m.Name = mOrginName
	curFields := []MessageField{}
	var filedTag int
	for _, field := range m.Fields {
		if isInSlice([]string{"create_time", "update_time", "version", "del_state", "delete_time"}, field.Name) {
			continue
		}
		filedTag++
		field.tag = filedTag
		field.Name = stringx.From(field.Name).ToCamelWithStartLower()
		if field.Comment == "" {
			field.Comment = field.Name
		}
		curFields = append(curFields, field)
	}
	m.Fields = curFields
	buf.WriteString(fmt.Sprintf("%s\n", m))

	//reset
	m.Name = mOrginName
	m.Fields = mOrginFields

	//resp
	// m.Name = "Update" + mOrginName + "Resp"
	// m.Fields = []MessageField{}
	// buf.WriteString(fmt.Sprintf("%s\n", m))

	// //reset
	// m.Name = mOrginName
	// m.Fields = mOrginFields
}

//gen add resp message
func (m Message) GenRpcDelReqMessage(buf *bytes.Buffer) {
	mOrginName := m.Name
	mOrginFields := m.Fields

	// m.Name = mOrginName + "Filter"
	// m.Fields = []MessageField{
	// 	{Name: "id", Typ: "int64", tag: 1, Comment: "id"},
	// }
	// buf.WriteString(fmt.Sprintf("%s\n", m))

	// //reset
	// m.Name = mOrginName
	// m.Fields = mOrginFields

	//resp
	m.Name = "Del" + mOrginName + "Resp"
	m.Fields = []MessageField{}
	buf.WriteString(fmt.Sprintf("%s\n", m))

	//reset
	m.Name = mOrginName
	m.Fields = mOrginFields
}

//gen add resp message
func (m Message) GenRpcGetByIdReqMessage(buf *bytes.Buffer) {
	mOrginName := m.Name
	mOrginFields := m.Fields

	m.Name = mOrginName + "Filter"
	m.Fields = []MessageField{
		{Name: "id", Typ: "int64", tag: 1, Comment: "id"},
	}
	buf.WriteString(fmt.Sprintf("%s\n", m))

	//reset
	m.Name = mOrginName
	m.Fields = mOrginFields

	//resp
	// firstWord := strings.ToLower(string(m.Name[0]))
	// m.Name = mOrginName
	// m.Fields = []MessageField{
	// 	{Typ: mOrginName, Name: stringx.From(firstWord + mOrginName[1:]).ToCamelWithStartLower(), tag: 1, Comment: stringx.From(firstWord + mOrginName[1:]).ToCamelWithStartLower()},
	// }
	// buf.WriteString(fmt.Sprintf("%s\n", m))

	// //reset
	// m.Name = mOrginName
	// m.Fields = mOrginFields
}

//gen add resp message
func (m Message) GenRpcSearchReqMessage(buf *bytes.Buffer, needList bool) {
	mOrginName := m.Name
	mOrginFields := m.Fields

	m.Name = mOrginName + "Filter"
	curFields := []MessageField{
		{Typ: "int64", Name: "pageNo", tag: 1, Comment: "pageNo"},
		{Typ: "int64", Name: "pageSize", tag: 2, Comment: "pageSize"},
	}
	var filedTag = len(curFields)
	for _, field := range m.Fields {
		if isInSlice([]string{"version", "del_state", "delete_time"}, field.Name) {
			continue
		}
		filedTag++
		field.tag = filedTag
		field.Name = stringx.From(field.Name).ToCamelWithStartLower()
		if field.Comment == "" {
			field.Comment = field.Name
		}
		curFields = append(curFields, field)
	}
	m.Fields = curFields

	tmpStr := fmt.Sprintf("%s\n", m)

	//增加数据库字段开始结束标签, 自定义标签
	tmpStr = strings.Replace(tmpStr, "{", "{\n  //Database Tag Begin. DO NOT EDIT!!! ", 1)
	tmpStr = strings.Replace(tmpStr, "}", "  //Database Tag End. DO NOT EDIT!!!  \n\n  //Custom Tag .You Can Edit. \n\n}", 1)

	buf.WriteString(tmpStr)

	//reset
	m.Name = mOrginName
	m.Fields = mOrginFields

	if needList {
		//resp
		firstWord := strings.ToLower(string(m.Name[0]))
		m.Name = mOrginName + "List"
		m.Fields = []MessageField{
			{Typ: "repeated " + mOrginName, Name: stringx.From(firstWord + mOrginName[1:]).ToCamelWithStartLower(), tag: 1, Comment: stringx.From(firstWord+mOrginName[1:]).ToCamelWithStartLower() + "List"},
			{Typ: "int64", Name: "totalPage", tag: 2},
			{Typ: "int64", Name: "totalCount", tag: 3},
			{Typ: "int64", Name: "curPage", tag: 4},
		}
		buf.WriteString(fmt.Sprintf("%s", m))
	}

	//reset
	m.Name = mOrginName
	m.Fields = mOrginFields
}

// String returns a string representation of a Message.
func (m Message) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("message %s {\n", m.Name))
	for _, f := range m.Fields {
		buf.WriteString(fmt.Sprintf("%s%s; //%s\n", indent, f, f.Comment))
	}
	buf.WriteString("}\n")

	return buf.String()
}

// AppendField appends a message field to a message. If the tag of the message field is in use, an error will be returned.
func (m *Message) AppendField(mf MessageField) error {
	for _, f := range m.Fields {
		if f.Tag() == mf.Tag() {
			return fmt.Errorf("tag `%d` is already in use by field `%s`", mf.Tag(), f.Name)
		}
	}

	m.Fields = append(m.Fields, mf)

	return nil
}

// MessageField represents the field of a message.
type MessageField struct {
	Typ     string
	Name    string
	tag     int
	Comment string
}

// NewMessageField creates a new message field.
func NewMessageField(typ, name string, tag int, comment string) MessageField {
	return MessageField{typ, name, tag, comment}
}

// Tag returns the unique numbered tag of the message field.
func (f MessageField) Tag() int {
	return f.tag
}

// String returns a string representation of a message field.
func (f MessageField) String() string {
	return fmt.Sprintf("%s %s = %d", f.Typ, f.Name, f.tag)
}

// Column represents a database column.
type Column struct {
	TableName              string
	TableComment           string
	ColumnName             string
	IsNullable             string
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	ColumnType             string
	ColumnComment          string
}

// Table represents a database table.
type Table struct {
	TableName  string
	ColumnName string
}

// parseColumn parses a column and inserts the relevant fields in the Message. If an enumerated type is encountered, an Enum will
// be added to the Schema. Returns an error if an incompatible protobuf data type cannot be found for the database column type.
func parseColumn(s *Schema, msg *Message, col Column) error {
	typ := strings.ToLower(col.DataType)
	var fieldType string

	switch typ {
	case "char", "varchar", "text", "longtext", "mediumtext", "tinytext":
		fieldType = "string"
	case "enum", "set":
		// Parse c.ColumnType to get the enum list
		enumList := regexp.MustCompile(`[enum|set]\((.+?)\)`).FindStringSubmatch(col.ColumnType)
		enums := strings.FieldsFunc(enumList[1], func(c rune) bool {
			cs := string(c)
			return "," == cs || "'" == cs
		})

		enumName := inflect.Singularize(snaker.SnakeToCamel(col.TableName)) + snaker.SnakeToCamel(col.ColumnName)
		enum, err := newEnumFromStrings(enumName, col.ColumnComment, enums)
		if nil != err {
			return err
		}

		s.Enums = append(s.Enums, enum)

		fieldType = enumName
	case "blob", "mediumblob", "longblob", "varbinary", "binary":
		fieldType = "bytes"
	case "date", "time", "datetime", "timestamp":
		//s.AppendImport("google/protobuf/timestamp.proto")
		fieldType = "int64"
	case "bool":
		fieldType = "bool"
	case "tinyint", "smallint", "int", "mediumint", "bigint":
		fieldType = "int64"
	case "float", "double":
		fieldType = "double"
	case "decimal":
		fieldType = "string"
	}

	if "" == fieldType {
		return fmt.Errorf("no compatible protobuf type found for `%s`. column: `%s`.`%s`", col.DataType, col.TableName, col.ColumnName)
	}

	field := NewMessageField(fieldType, col.ColumnName, len(msg.Fields)+1, col.ColumnComment)

	err := msg.AppendField(field)
	if nil != err {
		return err
	}

	return nil
}

func isInSlice(slice []string, s string) bool {
	for i := range slice {
		if slice[i] == s {
			return true
		}
	}
	return false
}
