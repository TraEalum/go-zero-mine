package apigen

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func apigen(_ *cobra.Command, _ []string) error {
	port := intPort
	password := stringPassword
	dir := stringDir
	host := stringHost
	user := stringUser
	schema := stringSchema
	table := stringTable
	serviceName := stringServiceName
	ignoreTableStr := stringIgnoreTables
	protoFile := stringProtoFile
	flag.Parse()

	if password != "" && schema != "" && table != "" && protoFile != "" {

		connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, schema)
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		ignoreTables := strings.Split(ignoreTableStr, ",")

		s, err := GenerateSchema(db, table, ignoreTables, serviceName, dir)
		if nil != err {
			log.Fatal(err)
		}

		s, err = GenerateProtoType(s, serviceName, protoFile)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}


	} else if protoFile != "" && serviceName != "" {
		s, err := GenerateProtoType(nil, serviceName, protoFile)

		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}

	} else {
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, schema)
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		ignoreTables := strings.Split(ignoreTableStr, ",")

		s, err := GenerateSchema(db, table, ignoreTables, serviceName, dir)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}
	}



	return nil
}

// 指定proto文件生成xxxParam.api中type
func GenerateProtoType(s *Schema, serviceName string, protoFile string) (*Schema, error) {
	dir := "api/desc/"
	var err error

	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Println(fmt.Sprintf("创建文件错误:%v", err))
			panic(err)
		}
	}

	if s == nil {
		s = &Schema{
			Dir: dir,
		}
	}


	s.Syntax = synatx
	s.ServiceName = serviceName


	if err = typesFromProto(s, protoFile, serviceName); err != nil {
		fmt.Println(err)
	}

	sort.Sort(s.Imports)
	sort.Sort(s.Messages)
	sort.Sort(s.Enums)

	return s, nil
}