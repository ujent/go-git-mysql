package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//var schema = "CREATE TABLE `files2` (`id` BIGINT AUTO_INCREMENT NOT NULL PRIMARY KEY, `ParentID` BIGINT,`name` varchar(255) NOT NULL, `path` varchar(255) NOT NULL, `position` bigint, `flag` TINYINT, `mode` TINYINT, `content` BINARY)"

var schema = "CREATE TABLE `files4` (`id` BIGINT AUTO_INCREMENT NOT NULL PRIMARY KEY, `parentID` BIGINT,`name` varchar(255) NOT NULL, `path` varchar(255) NOT NULL, `flag` TINYINT, `mode` TINYINT, `content` LONGBLOB)"

func main() {

	db, err := sqlx.Connect("mysql", "root:secret@/gogit")
	if err != nil {
		log.Fatalf("fatal error during opening db: %s", err)
	}
	defer db.Close()

	//db.MustExec(schema)

	parentID := int64(0)
	err = db.Get(&parentID, fmt.Sprintf("SELECT id FROM %s WHERE path=?", "files4"), "path/file.txt")

	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatal("empty")
		} else {
			log.Fatal(err)
		}
	}

	fmt.Print(parentID)

	//TODO - insert table "files"
	//db.MustExec(schema)

	// f := &mysqlfs.FileDB{
	// 	Name:    "file.txt",
	// 	Path:    "path/file.txt",
	// 	Content: []byte{},
	// 	Mode:    2,
	// 	Flag:    1,
	// }

	// stmt, err := db.Prepare(fmt.Sprintf("INSERT INTO %s(name,path,mode,flag, content) VALUES(?,?,?,?,?)", "files4"))

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer stmt.Close()

	// _, err = stmt.Exec(f.Name, f.Path, f.Mode, f.Flag, f.Content)

	// if err != nil {
	// 	log.Fatal(err)
	// }
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
