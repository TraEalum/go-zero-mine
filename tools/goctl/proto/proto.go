package proto

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

// Action provides the entry for goctl mongo code generation.
func proto(_ *cobra.Command, _ []string) error {
	port := intPort
	password := stringPassword
	dir := stringDir
	host := stringHost
	user := stringUser
	schema := stringSchema
	table := stringTable
	serviceName := stringServiceName
	packageName := stringPackage
	goPackageName := stringGoPackage
	ignoreTableStr := stringIgnoreTables

	flag.Parse()
	//fmt.Println(port)
	//return nil
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, schema)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	ignoreTables := strings.Split(ignoreTableStr, ",")

	s, err := GenerateSchema(db, table,ignoreTables,serviceName, goPackageName, packageName, dir)

	if nil != err {
		log.Fatal(err)
	}

	if nil != s {
		fmt.Println(s)
	}

	return nil
}
