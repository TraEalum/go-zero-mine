package apigen

import (
	"github.com/spf13/cobra"
)

const (
	//update、insert、query、delete
	UPDATE string = "update"
	INSERT string = "insert"
	QUERY  string = "query"
	DELETE string = "delete"
)

var (
	intPort            int
	stringPassword     string
	stringDir          string // 读取api文件的路径
	stringHost         string
	stringUser         string
	stringSchema       string // 库名
	stringTable        string // 表名
	stringServiceName  string // 服务名
	stringIgnoreTables string
	stringProtoFile    string
	stringCurdMethod   string

	// Cmd describes a model command.
	Cmd = &cobra.Command{
		Use:   "apigen",
		Short: "Generate model code",
		RunE:  apigen,
	}
)

func init() {
	Cmd.Flags().StringVar(&stringPassword, "password", "", "the database password")
	Cmd.Flags().StringVar(&stringDir, "dir", "api/desc/", "The target dir")
	Cmd.Flags().StringVar(&stringHost, "host", "localhost", "the database host")
	Cmd.Flags().IntVar(&intPort, "port", 3306, "the database port")
	Cmd.Flags().StringVar(&stringUser, "user", "root", "the database user")
	Cmd.Flags().StringVar(&stringSchema, "schema", "", "the database schema")
	Cmd.Flags().StringVar(&stringTable, "table", "", "the table schema，multiple tables ',' split. ")
	Cmd.Flags().StringVar(&stringServiceName, "serviceName", "", "the protobuf service name , defaults to the database schema.")
	Cmd.Flags().StringVar(&stringIgnoreTables, "ignore_tables", "", "a comma spaced list of tables to ignore")
	Cmd.Flags().StringVar(&stringProtoFile, "proto", "", "the proto file path")
	Cmd.Flags().StringVar(&stringCurdMethod, "i", "", "生成update、insert、query、delete方法，不传默认只生成查询方法,多个方法之间用逗号(,)分割")

}
