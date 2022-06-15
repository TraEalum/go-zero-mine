package apigen

import (
	"bufio"
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/chuckpreslar/inflect"
	"github.com/serenize/snaker"
	"github.com/zeromicro/go-zero/tools/goctl/util/stringx"
)

const (
	synatx = "v1"

	indent = "  "
)

func GenerateSchema(db *sql.DB, table string, ignoreTables []string, serviceName string, dir string) (*Schema, error) {
	s := &Schema{
		Dir: dir,
	}

	dbs, err := dbSchema(db)
	if nil != err {
		return nil, err
	}

	s.Syntax = synatx
	s.ServiceName = serviceName
	cols, err := dbColumns(db, dbs, table)
	if nil != err {
		return nil, err
	}
	//if the clos lenth equal 0
	if len(cols) == 0 {
		return nil, errors.New("no columns to genertor!!!")
	}
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
	if len(ignoreTables) != 0 {
		for _, ig := range ignoreTables {
			ignoreMap[ig] = true
		}
	}
	for _, c := range cols {
		if _, ok := ignoreMap[c.TableName]; ok {
			continue
		}

		messageName := snaker.SnakeToCamel(c.TableName)
		messageName = inflect.Singularize(messageName)

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

func dbColumns(db *sql.DB, schema, table string) ([]Column, error) {

	tableArr := strings.Split(table, ",")
	if len(tableArr) == 0 {
		return nil, errors.New("no table to genertor")
	}
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

		if cs.TableComment == "" {
			cs.TableComment = stringx.From(cs.TableName).ToCamelWithStartLower()
		}
		//这里过滤掉不需要生成的字段
		if !isInSlice([]string{"create_time", "update_time"}, cs.ColumnName) {
			cols = append(cols, cs)
		}

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
	Dir         string
	Imports     sort.StringSlice
	Messages    MessageCollection
	Enums       EnumCollection
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
	b := strings.HasSuffix(s.Dir, ".api")
	if !b {
		fmt.Println("Only API terminated files can be generated")
		panic("Only API terminated files can be generated")
	}
	//这里生成xxxParam.api文件 start
	arr := strings.Split(s.Dir, ".")
	paramFile := arr[0] + "Param." + arr[1]
	_, err := os.Stat(paramFile)
	if os.IsNotExist(err) {
		s.CreateParamString(paramFile)
	} else {
		s.UpdateParamString(paramFile)
	}
	_, err = os.Stat(s.Dir)
	//如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
	if os.IsNotExist(err) {
		return s.CreateString()
	}

	return s.UpdateString()
}
func (s *Schema) CreateParamString(fileName string) string {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("syntax = \"%s\";\n", s.Syntax))
	buf.WriteString("\n")
	buf.WriteString("// Already Exist Table:\n")
	for _, m := range s.Messages {
		buf.WriteString("// " + m.Name)
		buf.WriteString("\n")
	}
	buf.WriteString("// Exist Table End\n")
	buf.WriteString("\n")
	buf.WriteString("// Type Record Start\n")

	for _, m := range s.Messages {
		buf.WriteString("//--------------------------------" + m.Comment + "--------------------------------")
		buf.WriteString("\n")
		// 创建
		m.GenApiDefault(buf)
		m.GenApiDefaultResp(buf)
		//更新
		m.GenApiUpdateReq(buf)
		m.GenApiUpdateResp(buf)
		//查询
		m.GenApiQueryReq(buf)
		m.GenApiQueryResp(buf)

	}
	buf.WriteString("// Type Record End\n")
	err := ioutil.WriteFile(fileName, []byte(buf.String()), 0666)
	if err != nil {
		fmt.Println(fmt.Sprintf("生成paramFile文件失败:%s", err.Error()))
		return ""
	}
	return "paramFile Done"
}
func (s *Schema) UpdateParamString(fileName string) string {
	bufNew := new(bytes.Buffer)
	file, err := os.OpenFile(fileName, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Sprintf("Open file error!%v", err)
	}
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
		existTableName = append(existTableName, line[3:])
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
		if strings.Contains(line, "Type Record End") {
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
		if !isInSlice(existTableName, m.Name) {
			bufNew.WriteString("//--------------------------------" + m.Comment + "--------------------------------")
			bufNew.WriteString("\n")

			// 创建
			m.GenApiDefault(bufNew)
			m.GenApiDefaultResp(bufNew)
			//更新
			m.GenApiUpdateReq(bufNew)
			m.GenApiUpdateResp(bufNew)
			//查询
			m.GenApiQueryReq(bufNew)
			m.GenApiQueryResp(bufNew)

		}
	}
	bufNew.WriteString("// Type Record End\n")
	err = ioutil.WriteFile(fileName, []byte(bufNew.String()), 0666)
	return "paramFile DONE"
}

// String returns a string representation of a Schema.
func (s *Schema) CreateString() string {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("syntax = \"%s\";\n", s.Syntax))
	buf.WriteString("\n")
	buf.WriteString("import (\n")
	arr := strings.Split(s.Dir, ".")
	buf.WriteString(fmt.Sprintf("\t\"%s\" \n", arr[0]+"Param."+arr[1]))
	buf.WriteString(")")
	buf.WriteString("\n")
	buf.WriteString("// Already Exist Table:\n")
	for _, m := range s.Messages {
		buf.WriteString("// " + m.Name)
		buf.WriteString("\n")
	}
	buf.WriteString("// Exist Table End\n")
	buf.WriteString("\n")
	if len(s.Enums) > 0 {
		buf.WriteString("// Enums Record Start\n")
		for _, e := range s.Enums {
			buf.WriteString(fmt.Sprintf("%s\n", e))
		}
		buf.WriteString("// Enums Record End\n")
	}

	buf.WriteString("\n")
	buf.WriteString("// ------------------------------------ \n")
	buf.WriteString("// api Func\n")
	buf.WriteString("// ------------------------------------ \n\n")

	funcTpl := "service " + s.ServiceName + "{\n"
	for _, m := range s.Messages {
		funcTpl += "\t//-----------------------" + m.Comment + "----------------------- \n"
		firstUpperName := FirstUpper(m.Name)
		funcTpl += "\t@doc  创建" + m.Name + "\n"
		funcTpl += "\t@handler  create" + m.Name + "\n"
		funcTpl += "\tpost /" + stringx.From(m.Name).ToSnake() + "/create" + firstUpperName + " (" + m.Name + ") returns (Create" + firstUpperName + "Resp); \n\n"

		funcTpl += "\t@doc  更新" + m.Name + "\n"
		funcTpl += "\t@handler  update" + m.Name + "\n"
		funcTpl += "\tpost /" + stringx.From(m.Name).ToSnake() + "/update" + firstUpperName + " (Update" + m.Name + "Req) returns (Update" + firstUpperName + "Resp); \n\n"

		funcTpl += "\t@doc  查找" + m.Name + "\n"
		funcTpl += "\t@handler  query" + m.Name + "\n"
		funcTpl += "\tget /" + stringx.From(m.Name).ToSnake() + "/query" + firstUpperName + " (Query" + firstUpperName + "Req) returns (Query" + firstUpperName + "Resp); \n\n"

	}
	funcTpl = funcTpl + "\t // Service Record End\n"
	funcTpl = funcTpl + "}"
	buf.WriteString(funcTpl)
	err := ioutil.WriteFile(s.Dir, []byte(buf.String()), 0666)
	if err != nil {
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
		existTableName = append(existTableName, line[3:])
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

		if strings.Contains(line, "api Func") {
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

	// 写api接口名
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
			funcTpl += "\t//-----------------------" + m.Comment + "----------------------- \n"

			firstUpperName := FirstUpper(m.Name)
			funcTpl += "\t@doc  创建" + m.Name + "\n"
			funcTpl += "\t@handler  create" + m.Name + "\n"
			funcTpl += "\tpost /" + stringx.From(m.Name).ToSnake() + "/create" + firstUpperName + " (" + m.Name + ") returns (Create" + firstUpperName + "Resp); \n\n"

			funcTpl += "\t@doc  更新" + m.Name + "\n"
			funcTpl += "\t@handler  update" + m.Name + "\n"
			funcTpl += "\tpost /" + stringx.From(m.Name).ToSnake() + "/update" + firstUpperName + " (Update" + m.Name + "Req) returns (Update" + firstUpperName + "Resp); \n\n"

			funcTpl += "\t@doc  查找" + m.Name + "\n"
			funcTpl += "\t@handler  query" + m.Name + "\n"
			funcTpl += "\tget /" + stringx.From(m.Name).ToSnake() + "/query" + firstUpperName + " (Query" + firstUpperName + "Req) returns (Query" + firstUpperName + "Resp); \n\n"
		}
	}
	funcTpl = funcTpl + "\t // Service Record End\n"
	funcTpl = funcTpl + "}"

	bufNew.WriteString(funcTpl)
	_ = ioutil.WriteFile(s.Dir, []byte(bufNew.String()), 0666) //写入文件(字节数组)
	return "Done"
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

// Message represents a protocol buffer message.
type Message struct {
	Name    string
	Comment string
	Fields  []MessageField
}

//gen default message
func (m Message) GenApiDefault(buf *bytes.Buffer) {
	mOrginName := m.Name
	mOrginFields := m.Fields
	curFields := []MessageField{}
	for _, field := range m.Fields {
		if isInSlice([]string{"version", "del_state", "delete_time"}, field.Name) {
			continue
		}
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
}

// 先固定写为id
func (m Message) GenApiDefaultResp(buf *bytes.Buffer) {
	mOrginName := FirstUpper(m.Name)
	buf.WriteString(fmt.Sprintf("type Create%sResp {\n", mOrginName))
	buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s\"`   \n", indent, "Id", "int64", "id"))
	buf.WriteString("}\n")
}

func (m Message) GenApiUpdateReq(buf *bytes.Buffer) {
	mOrginName := FirstUpper(m.Name)
	buf.WriteString(fmt.Sprintf("type Update%sReq {\n", mOrginName))
	for _, f := range m.Fields {
		buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s\"`   //%s\n", indent, f.Name, f.Typ, f.ColumnName, f.Comment))
	}
	buf.WriteString("}\n")
}
func (m Message) GenApiUpdateResp(buf *bytes.Buffer) {
	mOrginName := FirstUpper(m.Name)
	buf.WriteString(fmt.Sprintf("type Update%sResp {\n", mOrginName))
	buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s\"`   \n", indent, "Id", "int64", "id"))
	buf.WriteString("}\n")
}

//先固定三个参数
func (m Message) GenApiQueryReq(buf *bytes.Buffer) {
	mOrginName := FirstUpper(m.Name)
	buf.WriteString(fmt.Sprintf("type Query%sReq {\n", mOrginName))
	buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s,optional\"`   \n", indent, "Id", "int64", "id"))
	buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s,optional\"`   \n", indent, "PageNo", "int64", "page_no"))
	buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s,optional\"`   \n", indent, "PageSize", "int64", "page_size"))
	buf.WriteString("}\n")
}

func (m Message) GenApiQueryResp(buf *bytes.Buffer) {
	mOrginName := FirstUpper(m.Name)
	buf.WriteString(fmt.Sprintf("type Query%sResp {\n", mOrginName))
	buf.WriteString(fmt.Sprintf("%s  %s   []%s  `json:\"%s\"`   \n", indent, "Data", mOrginName, "data"))
	buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s\"`   \n", indent, "CurrPage", "int64", "curr_page"))
	buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s\"`   \n", indent, "TotalPage", "int64", "total_page"))
	buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s\"`   \n", indent, "TotalCount", "int64", "total_count"))
	buf.WriteString("}\n")
}

// String returns a string representation of a Message.
func (m Message) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("type %s {\n", m.Name))
	for _, f := range m.Fields {
		buf.WriteString(fmt.Sprintf("%s  %s   %s  `json:\"%s\"`  ; //%s\n", indent, FirstUpper(f.Name), f.Typ, f.ColumnName, f.Comment))
	}
	buf.WriteString("}\n")

	return buf.String()
}

// MessageField represents the field of a message.
type MessageField struct {
	Typ        string
	Name       string
	Comment    string
	ColumnName string
}

// NewMessageField creates a new message field.
func NewMessageField(typ, name, comment, columnName string) MessageField {
	return MessageField{typ, name, comment, columnName}
}

func (m *Message) AppendField(mf MessageField) error {
	m.Fields = append(m.Fields, mf)
	return nil
}

// String returns a string representation of a message field.
func (f MessageField) String() string {
	return fmt.Sprintf("%s %s  `json:\"%s\"`", f.Name, f.Typ, f.ColumnName)
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
	case "float", "decimal", "double":
		fieldType = "double"
	}

	if "" == fieldType {
		return fmt.Errorf("no compatible go type found for `%s`. column: `%s`.`%s`", col.DataType, col.TableName, col.ColumnName)
	}
	field := NewMessageField(fieldType, col.ColumnName, col.ColumnComment, col.ColumnName)
	err := msg.AppendField(field)
	if nil != err {
		return err
	}

	return nil
}

func isInSlice(slice []string, s string) bool {
	for i := range slice {
		if strings.TrimSpace(slice[i]) == strings.TrimSpace(s) {
			return true
		}
	}
	return false
}
func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
