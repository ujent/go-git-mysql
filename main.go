package main

import (
	_ "database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/ujent/go-git-mysql/mysqlfs"
)

//var schema = "CREATE TABLE `files2` (`id` BIGINT AUTO_INCREMENT NOT NULL PRIMARY KEY, `ParentID` BIGINT,`name` varchar(255) NOT NULL, `path` varchar(255) NOT NULL, `position` bigint, `flag` TINYINT, `mode` TINYINT, `content` BINARY)"

var schema = "CREATE TABLE `files3` (`id` BIGINT AUTO_INCREMENT NOT NULL PRIMARY KEY, `ParentID` BIGINT,`name` varchar(255) NOT NULL, `path` varchar(255) NOT NULL, `position` bigint, `flag` TINYINT, `mode` TINYINT, `content` LONGBLOB)"

func main() {

	db, err := sqlx.Connect("mysql", "root:secret@/gogit")
	if err != nil {
		log.Fatalf("fatal error during opening db: %s", err)
	}
	defer db.Close()

	//TODO - insert table "files"
	//db.MustExec(schema)

	f := &mysqlfs.FileDB{
		Name:    "file.txt",
		Path:    "path/file.txt",
		Content: []byte{},
		Mode:    2,
		Flag:    1,
	}

	//log.Println(fmt.Sprintf("INSERT INTO %s(name,path,content,mode,flag) VALUES(%s,%s,%d,%d,%d)", "files2", f.Name, f.Path, f.Content, f.Mode, f.Flag))
	stmtIns, err := db.Prepare(fmt.Sprintf("INSERT INTO %s(name,path,mode,flag, content) VALUES(?,?,?,?,?)", "files3"))
	if err != nil {
		log.Fatal(err)
	}

	defer stmtIns.Close()

	_, err = stmtIns.Exec(f.Name, f.Path, f.Mode, f.Flag, f.Content)

	if err != nil {
		log.Fatal(err)
	}
	// query

	// rows, err := db.Query("select id, name from test")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// type row struct {
	// 	id       int
	// 	name     string
	// 	parentid sql.NullInt64
	// }

	// for rows.Next() {
	// 	r := row{}
	// 	err := rows.Scan(&r.id, &r.name)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	var parentid int64
	// 	if r.parentid.Valid {
	// 		parentid = r.parentid.Int64
	// 	}

	// 	log.Printf("id: %d, name: %s %d\n", r.id, r.name, parentid)
	// }
	// if err := rows.Err(); err != nil {
	// 	log.Fatal(err)
	// }

}
