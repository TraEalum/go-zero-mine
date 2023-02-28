package apigen

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"path"
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
	generateCurdMethod := stringCurdMethod

	generateCurdMethod = strings.TrimSpace(generateCurdMethod)
	generateMethod := strings.Split(generateCurdMethod, ",")

	flag.Parse()

	dir = path.Join(dir, "/") + "/"

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
		s.GenerateCurdMethod = generateMethod

		s, err = GenerateProtoType(s, serviceName, protoFile, dir)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}

	} else if protoFile != "" && serviceName != "" {
		s, err := GenerateProtoType(nil, serviceName, protoFile, dir)

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
		s.GenerateCurdMethod = generateMethod

		if nil != s {
			fmt.Println(s)
		}
	}

	return nil
}
