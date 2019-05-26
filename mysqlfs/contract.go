package mysqlfs

import (
	"os"
)

type Storage interface {
	HasFile(path string) (bool, error)
	NewFile(path string, mode os.FileMode, flag int) (*FileDB, error)
	GetFile(path string) (*FileDB, error)
	MustGetFile(path string) (*FileDB, error)

	UpdateFileContent(fileID int64, content []byte) error

	//AddFile(fileName string) error
	//GetFile(filename string) (billy.File, error)
}

//FileDB - main db obect for saving files
type FileDB struct { //TODO - really need all of this?
	ID       int64       `db:"id"`
	ParentID int64       `db:"parentID"`
	Name     string      `db:"name"`
	Path     string      `db:"path"`
	Content  []byte      `db:"content"`
	Position int64       `db:"position"`
	Flag     int         `db:"flag"`
	Mode     os.FileMode `db:"mode"`
	Size     int64       `db:"size"`
}

//File - Mysql fs object, realizes interface billy.File
type File struct {
	ID       int64
	ParentID int64
	FileName string
	Path     string
	Content  []byte
	Position int64
	Flag     int
	Mode     os.FileMode

	IsClosed bool
}

//ToDo - not necessary?
type Content struct {
	Bytes []byte
}

type FileInfo struct {
	FileID   int64
	FileName string
	FileSize int64
	FileMode os.FileMode
}
