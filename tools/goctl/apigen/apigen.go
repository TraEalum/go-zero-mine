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
	apiFile := stringApiFile
	generateCurdMethod := stringCurdMethod

	generateCurdMethod = strings.TrimSpace(generateCurdMethod)
	generateMethod := strings.Split(generateCurdMethod, ",")

	flag.Parse()

	dir = path.Join(dir, "/") + "/"

	if password != "" && schema != "" && table != "" && apiFile != "" {

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

		s, err = GenerateApiType(s, serviceName, apiFile, dir)
		if nil != err {
			log.Fatal(err)
		}

		if nil != s {
			fmt.Println(s)
		}

	} else if apiFile != "" && serviceName != "" {
		s, err := GenerateApiType(nil, serviceName, apiFile, dir)

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
