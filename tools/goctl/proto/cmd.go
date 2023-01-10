package proto

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
	intPort        int
	stringPassword string
	stringDir      string
	stringHost     string
	stringUser     string
	//stringDatabase string
	stringSchema       string
	stringTable        string
	stringServiceName  string
	stringPackage      string
	stringGoPackage    string
	stringIgnoreTables string

	//sub table
	stringSubTableKey string
	intSubTableNumber int

	//根据table的传参来决定是否生成curd方法
	stringCurdMethod string

	// Cmd describes a model command.
	Cmd = &cobra.Command{
		Use:   "proto",
		Short: "Generate model code",
		RunE:  proto,
	}
)

func init() {
	Cmd.Flags().StringVar(&stringPassword, "password", "", "the database password")
	Cmd.Flags().StringVar(&stringDir, "dir", "", "The target dir")
	Cmd.Flags().StringVar(&stringHost, "host", "localhost", "the database host")
	Cmd.Flags().IntVar(&intPort, "port", 3306, "the database port")
	Cmd.Flags().StringVar(&stringUser, "user", "root", "the database user")
	// Cmd.Flags().StringVar(&stringDatabase, "db", "d", "The name of database [optional]")
	Cmd.Flags().StringVar(&stringSchema, "schema", "", "the database schema")
	Cmd.Flags().StringVar(&stringTable, "table", "", "the table schema，multiple tables ',' split. ")
	Cmd.Flags().StringVar(&stringServiceName, "serviceName", "", "the protobuf service name , defaults to the database schema.")
	Cmd.Flags().StringVar(&stringPackage, "package", "", "the protocol buffer package. defaults to the database schema.")
	Cmd.Flags().StringVar(&stringGoPackage, "goPackage", "", "the protocol buffer go_package. defaults to the database schema.")
	Cmd.Flags().StringVar(&stringIgnoreTables, "ignore_tables", "", "a comma spaced list of tables to ignore")
	Cmd.Flags().IntVarP(&intSubTableNumber, "subTableNumber", "s", 0, "Sub table number")
	Cmd.Flags().StringVarP(&stringSubTableKey, "subTableKey", "k", "", "Sub table key")
	Cmd.Flags().StringVar(&stringCurdMethod, "i", "", "生成update、insert、query、delete方法，不传默认只生成查询方法,多个方法之间用逗号(,)分割")

	//Cmd.AddCommand(mysqlCmd)
	//Cmd.AddCommand(mongoCmd)
}
