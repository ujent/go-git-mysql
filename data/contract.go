package data

import (
	"os"

	"gopkg.in/src-d/go-billy.v4"
)

type Service interface {
	AddFile(fileName string) error
	GetFile(filename string) (billy.File, error)
}

type File struct { //TODO - really need all of this?
	ID       float64     `db:"id"`
	Name     string      `db:"name"`
	Content  *Content    `db:"content"`
	Position int64       `db:"position"`
	Flag     int         `db:"flag"`
	Mode     os.FileMode `db:mode`

	IsClosed bool
}

type Content struct {
	Bytes []byte
}

type FileInfo struct {
	Name string      `db:"name"`
	Size int         `db:"size"`
	Mode os.FileMode `db:"mode"`
}
