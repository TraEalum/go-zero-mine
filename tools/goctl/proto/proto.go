package proto

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func reRootDir(dir string) string {
	// re_root_dir= ./../../../go_service/
	re_root_dir := os.Getenv("re_root_dir")
	if re_root_dir == "" {
		return dir
	}
	dir = re_root_dir + dir
	curr_dir, _ := os.Getwd()
	dir, _ = filepath.Abs(curr_dir + dir)
	// log.Println("re dir:", dir)
	return dir
}

// Action provides the entry for goctl mongo code generation.
func proto(_ *cobra.Command, _ []string) error {
	port := intPort
	password := stringPassword
	dir := stringDir
	dir = reRootDir(dir)
	host := stringHost
	user := stringUser
	schema := stringSchema
	table := stringTable
	serviceName := stringServiceName
	packageName := stringPackage
	goPackageName := stringGoPackage
	ignoreTableStr := stringIgnoreTables
	subTableKey := stringSubTableKey
	subTableNumber := intSubTableNumber
	generateCurdMethod := stringCurdMethod
	flag.Parse()

	if err := ifNotExistThenCreate(dir); err != nil {
		log.Fatal(err)
	}

	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, schema)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	ignoreTables := strings.Split(ignoreTableStr, ",")

	generateCurdMethod = strings.TrimSpace(generateCurdMethod)
	generateMethod := strings.Split(generateCurdMethod, ",")

	s, err := GenerateSchema(db, table, ignoreTables, serviceName, goPackageName, packageName, dir, subTableNumber, subTableKey, generateMethod)

	if nil != err {
		log.Fatal(err)
	}

	if nil != s {
		fmt.Println(s)
	}

	return nil
}

func ifNotExistThenCreate(path string) (err error) {
	split := strings.Split(path, "/")
	path = ""
	for i := 0; i < len(split)-1; i++ {
		path = filepath.Join(path, split[i])
	}

	if _, err = os.Stat(path); err != nil {
		return os.MkdirAll(path, os.ModePerm)
	}

	return nil
}
