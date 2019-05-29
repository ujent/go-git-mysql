package mysqlfs

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const separator = filepath.Separator
const fileTableName = "file"

type storage struct {
	db *sqlx.DB
}

func newStorage(connectionStr string) (Storage, error) {
	db, err := sqlx.Connect("mysql", connectionStr)

	if err != nil {
		return nil, err
	}

	return &storage{db: db}, nil
}

func (s *storage) GetFile(path string) (*File, error) {
	path = clean(path)
	f := FileDB{}

	err := s.db.Get(&f, fmt.Sprintf("SELECT * FROM %s WHERE path = ?", fileTableName), path)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return fileDBtoFile(&f), nil
}

func (s *storage) GetFileID(path string) (int64, error) {
	path = clean(path)
	id := int64(0)

	err := s.db.Get(&id, fmt.Sprintf("SELECT id FROM %s WHERE path = ?", fileTableName), path)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0, err
	}

	return id, nil
}

func (s *storage) NewFile(path string, mode os.FileMode, flag int) (*File, error) {
	path = clean(path)
	f, err := s.GetFile(path)

	if err != nil {
		return nil, err
	}

	if f != nil {
		if !f.Mode.IsDir() {
			return nil, fmt.Errorf("file already exists %q", path)
		}

		return nil, nil
	}

	f = &File{
		FileName: filepath.Base(path),
		Path:     path,
		Content:  []byte{},
		Mode:     mode,
		Flag:     flag,
		storage:  s,
	}

	stmtIns, err := s.db.Prepare(fmt.Sprintf("INSERT INTO %s(name,path,mode,flag, content) VALUES(?,?,?,?,?)", fileTableName))
	if err != nil {
		return nil, err
	}

	defer stmtIns.Close()

	res, err := stmtIns.Exec(f.FileName, f.Path, f.Mode, f.Flag, f.Content)

	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	f.ID = id
	s.createParentAddToFile(path, mode, f)

	return f, nil
}

func (s *storage) Children(path string) ([]*File, error) {
	path = clean(path)
	parentID := int64(0)

	err := s.db.Get(&parentID, fmt.Sprintf("SELECT id FROM %s WHERE path=?", fileTableName), path)

	if err != nil {
		if err == sql.ErrNoRows {
			return []*File{}, nil
		}
		return nil, err
	}

	return s.ChildrenByFileID(parentID)
}

func (s *storage) ChildrenIds(path string) ([]int64, error) {
	path = clean(path)
	parentID := int64(0)

	err := s.db.Get(&parentID, fmt.Sprintf("SELECT id FROM %s WHERE path=?", fileTableName), path)

	if err != nil {
		if err == sql.ErrNoRows {
			return []int64{}, nil
		}
		return nil, err
	}

	return s.ChildrenIdsByFileID(parentID)
}

func (s *storage) ChildrenByFileID(id int64) ([]*File, error) {
	resDB := []FileDB{}
	err := s.db.Select(&resDB, fmt.Sprintf("SELECT * FROM %s WHERE parentID=?", fileTableName), id)

	if err != nil {
		return nil, err
	}

	res := make([]*File, 0)
	for _, fDB := range resDB {
		f := fileDBtoFile(&fDB)
		res = append(res, f)
	}

	return res, nil
}

func (s *storage) ChildrenIdsByFileID(id int64) ([]int64, error) {
	res := []int64{}
	err := s.db.Select(&res, fmt.Sprintf("SELECT id FROM %s WHERE parentID=?", fileTableName), id)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *storage) RenameFile(from, to string) error {
	from = clean(from)
	to = clean(to)
	f, err := s.GetFile(from)

	if err != nil {
		return err
	}

	if f == nil {
		return os.ErrNotExist
	}

	newName := filepath.Base(to)

	if f.Mode.IsDir() {

		children, err := s.ChildrenByFileID(f.ID)
		if err != nil {
			return err
		}

		tx := s.db.MustBegin()
		tx.MustExec(fmt.Sprintf("UPDATE %s SET name=?, path=? WHERE id=?", fileTableName), newName, to)

		if len(children) != 0 {
			for _, c := range children {
				tx.MustExec(fmt.Sprintf("UPDATE %s SET path=? WHERE id=?", fileTableName), filepath.Join(to, c.FileName), c.ID)
			}
		}

		err = tx.Commit()

		if err != nil {
			return err
		}

	} else {
		newParentID, err := s.GetFileID(filepath.Dir(to))

		if err != nil {
			return err
		}

		if newParentID != 0 {
			stmt, err := s.db.Prepare(fmt.Sprintf("UPDATE %s SET name=?, path=?, parentID=? WHERE id=?", fileTableName))

			if err != nil {
				return err
			}

			defer stmt.Close()

			_, err = stmt.Exec(newName, to, newParentID, f.ID)

			if err != nil {
				return err
			}

			return nil
		}

		newParent, err := s.createParent(to, 0644, f)

		if err != nil {
			return err
		}

		if newParent == nil {
			stmt, err := s.db.Prepare(fmt.Sprintf("UPDATE %s SET name=?, path=?, parentID=? WHERE id=?", fileTableName))

			if err != nil {
				return err
			}

			defer stmt.Close()

			_, err = stmt.Exec(newName, to, nil, f.ID)
		} else {
			stmt, err := s.db.Prepare(fmt.Sprintf("UPDATE %s SET name=?, path=?, parentID=? WHERE id=?", fileTableName))

			if err != nil {
				return err
			}

			defer stmt.Close()

			_, err = stmt.Exec(newName, to, newParent.ID, f.ID)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *storage) RemoveFile(path string) error {
	path = clean(path)

	f, err := s.GetFile(path)

	if err != nil {
		return err
	}

	if f == nil {
		return os.ErrNotExist
	}

	childrenIds, err := s.ChildrenIdsByFileID(f.ID)

	if err != nil {
		return err
	}

	childrenIdsLen := len(childrenIds)

	if f.Mode.IsDir() && childrenIdsLen != 0 {
		return fmt.Errorf("dir: %s contains files", path)
	}

	stmt, err := s.db.Prepare(fmt.Sprintf("DELETE FROM %s where id=?", fileTableName))
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(f.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *storage) UpdateFileContent(fileID int64, content []byte) error {
	stmt, err := s.db.Prepare(fmt.Sprintf("UPDATE %s SET content=? WHERE id=?", fileTableName))
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(content, fileID)

	if err != nil {
		return err
	}

	return nil
}

func (s *storage) createParent(path string, mode os.FileMode, f *File) (*File, error) {
	base := filepath.Dir(path)
	base = clean(base)

	if f.FileName == string(separator) {
		return nil, nil
	}

	parent, err := s.NewFile(base, mode.Perm()|os.ModeDir, 0)
	if err != nil {
		return nil, err
	}

	return parent, nil
}

func (s *storage) createParentAddToFile(path string, mode os.FileMode, f *File) error {
	parent, err := s.createParent(path, mode, f)

	if err != nil {
		return err
	}

	if parent == nil {
		return nil
	}

	f.ParentID = parent.ID

	stmt, err := s.db.Prepare(fmt.Sprintf("UPDATE %s SET parentID=? WHERE id=?", fileTableName))
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(parent.ID, f.ID)

	if err != nil {
		return err
	}

	return nil
}

func fileDBtoFile(f *FileDB) *File {
	if f == nil {
		return nil
	}

	var parID int64
	if f.ParentID.Valid {
		parID = f.ParentID.Int64
	}

	return &File{
		ID:       f.ID,
		FileName: f.Name,
		ParentID: parID,
		Path:     f.Path,
		Content:  f.Content,
		Flag:     f.Flag,
		Mode:     f.Mode,
	}
}

func clean(path string) string {
	return filepath.Clean(filepath.FromSlash(path))
}
