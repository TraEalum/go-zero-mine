package proto

import (
	"github.com/spf13/cobra"
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

	//Cmd.AddCommand(mysqlCmd)
	//Cmd.AddCommand(mongoCmd)
}
