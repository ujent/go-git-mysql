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

	files    map[string]*FileDB
	children map[string]map[string]*FileDB
}

func newStorage(connectionStr string) (*storage, error) {
	db, err := sqlx.Connect("mysql", connectionStr)

	if err != nil {
		return nil, err
	}

	return &storage{db: db}, nil
}

func (s *storage) HasFile(path string) (bool, error) {
	path = clean(path)

	err := s.db.QueryRow(fmt.Sprintf("select p.id from %s as p where p.path = %s;", fileTableName, path)).Scan()

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *storage) GetFile(path string) (*FileDB, error) {
	path = clean(path)
	f := FileDB{}

	err := s.db.Get(&f, "select p from $1 as p where p.path = $2", fileTableName, path)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &f, nil
}

func (s *storage) NewFile(path string, mode os.FileMode, flag int) (*FileDB, error) {
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

	f = &FileDB{
		Name:    filepath.Base(path),
		Path:    path,
		Content: []byte{},
		Mode:    mode,
		Flag:    flag,
	}

	//ToDo create parent (?)
	stmtIns, err := s.db.Prepare(fmt.Sprintf("INSERT INTO %s(name,path,mode,flag, content) VALUES(?,?,?,?,?)", fileTableName))
	if err != nil {
		return nil, err
	}

	defer stmtIns.Close()

	_, err = stmtIns.Exec(f.Name, f.Path, f.Mode, f.Flag, f.Content)

	if err != nil {
		return nil, err
	}

	s.createParent(path, mode, f)
	return f, nil
}

func (s *storage) createParent(path string, mode os.FileMode, f *FileDB) error {
	base := filepath.Dir(path)
	base = clean(base)
	if f.Name == string(separator) {
		return nil
	}

	if _, err := s.NewFile(base, mode.Perm()|os.ModeDir, 0); err != nil {
		return err
	}

	if _, ok := s.children[base]; !ok {
		s.children[base] = make(map[string]*FileDB, 0)
	}

	s.children[base][f.Name] = f
	return nil
}

func (s *storage) Children(path string) []*FileDB {
	path = clean(path)

	l := make([]*FileDB, 0)
	for _, f := range s.children[path] {
		l = append(l, f)
	}

	return l
}

func (s *storage) MustGetFile(path string) (*FileDB, error) {
	f, err := s.GetFile(path)

	if err != nil {
		return nil, err
	}

	return f, nil
}

func (s *storage) Get(path string) (*FileDB, bool, error) {
	path = clean(path)
	hasPath, err := s.HasFile(path)

	if err != nil {
		return nil, false, err
	}

	if !hasPath {
		return nil, false, nil
	}

	file, ok := s.files[path]
	return file, ok, nil
}

func (s *storage) Rename(from, to string) error {
	from = clean(from)
	to = clean(to)
	hasPath, err := s.HasFile(from)

	if err != nil {
		return err
	}

	if !hasPath {
		return os.ErrNotExist
	}

	move := [][2]string{{from, to}}

	for pathFrom := range s.files {
		if pathFrom == from || !filepath.HasPrefix(pathFrom, from) {
			continue
		}

		rel, _ := filepath.Rel(from, pathFrom)
		pathTo := filepath.Join(to, rel)

		move = append(move, [2]string{pathFrom, pathTo})
	}

	for _, ops := range move {
		from := ops[0]
		to := ops[1]

		if err := s.move(from, to); err != nil {
			return err
		}
	}

	return nil
}

func (s *storage) move(from, to string) error {
	s.files[to] = s.files[from]
	s.files[to].Name = filepath.Base(to)
	s.children[to] = s.children[from]

	defer func() {
		delete(s.children, from)
		delete(s.files, from)
		delete(s.children[filepath.Dir(from)], filepath.Base(from))
	}()

	return s.createParent(to, 0644, s.files[to])
}

func (s *storage) Remove(path string) error {
	path = clean(path)

	f, has, err := s.Get(path)

	if err != nil {
		return err
	}

	if !has {
		return os.ErrNotExist
	}

	if f.Mode.IsDir() && len(s.children[path]) != 0 {
		return fmt.Errorf("dir: %s contains files", path)
	}

	base, file := filepath.Split(path)
	base = filepath.Clean(base)

	delete(s.children[base], file)
	delete(s.files, path)
	return nil
}

func clean(path string) string {
	return filepath.Clean(filepath.FromSlash(path))
}

func (s *storage) UpdateFileContent(fileID int64, content []byte) error {
	return nil
}
