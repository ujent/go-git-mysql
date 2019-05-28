package mysqlfs

import (
	"database/sql"
	"os"
)

type Storage interface {
	NewFile(path string, mode os.FileMode, flag int) (*File, error)
	GetFile(path string) (*File, error)
	RenameFile(from, to string) error
	RemoveFile(path string) error
	Children(path string) ([]*File, error)
	ChildrenIdsByFileID(id int64) ([]int64, error)
	ChildrenByFileID(id int64) ([]*File, error)

	UpdateFileContent(fileID int64, content []byte) error
}

//FileDB - main db obect for saving files
type FileDB struct { //TODO - really need all of this?
	ID       int64         `db:"id"`
	ParentID sql.NullInt64 `db:"parentID"`
	Name     string        `db:"name"`
	Path     string        `db:"path"`
	Content  []byte        `db:"content"`
	Flag     int           `db:"flag"`
	Mode     os.FileMode   `db:"mode"`
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
	storage  *storage
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
