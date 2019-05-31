package mysqlfs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"bytes"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var schema = "CREATE TABLE `file` (`id` BIGINT AUTO_INCREMENT NOT NULL PRIMARY KEY, `parentID` BIGINT,`name` varchar(255) NOT NULL, `path` varchar(255) NOT NULL, `flag` INT, `mode` BIGINT, `content` LONGBLOB)"
var connStr = "root:secret@/gogit"

func TestNewStorage(t *testing.T) {
	_, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}
}

func TestCreate(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	fs, err := New(connStr)

	if err != nil {
		t.Error(err)
	}

	f, err := fs.Create("/dir1/dir2/file1.txt")

	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%s,", f.Name())

	fName := f.Name()

	if fName == "" {
		t.Errorf("Wrong file name: %s", fName)
	}

	dropTable(connStr)
}

func TestCreateParent(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	par, err := createParent(s, "dir1/dir2/file1.txt", 0644)

	if err != nil {
		t.Error(err)
	}

	if par == nil {
		t.Error("Parent wasn't created")
	}

	dropTable(connStr)

}

func TestOpenFile(t *testing.T) {
	err := createTable(connStr)

	fs, err := New(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"
	f, err := fs.Create(path)

	if err != nil {
		t.Error(err)
	}

	bf, err := fs.Open(path)

	if err != nil {
		t.Error(err)
	}

	if bf == nil {
		t.Error("No file")
	}

	if f.Name() != bf.Name() {
		t.Errorf("wrong name: created - %s, got - %s", f.Name(), bf.Name())
	}
	dropTable(connStr)
}

func TestStat(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	fs, err := New(connStr)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"

	f, err := s.NewFile(path, 0666, os.O_RDWR|os.O_CREATE|os.O_TRUNC)

	if err != nil {
		t.Error(err)
	}

	fi1, err := fs.Stat(path)

	if err != nil {
		t.Error(err)
	}

	fi2, err := f.Stat()

	if err != nil {
		t.Error(err)
	}

	if fi1.Name() != fi2.Name() {
		t.Errorf("fs fileName: %s, storage file name: %s", fi1.Name(), fi2.Name())
	}

	if fi1.Size() != fi2.Size() {
		t.Errorf("fs size: %d, storage size: %d", fi1.Size(), fi2.Size())
	}

	if fi1.Mode() != fi2.Mode() {
		t.Errorf("fs Mode: %d, storage Mode: %d", fi1.Mode(), fi2.Mode())
	}

	dropTable(connStr)
}

func TestTempFile(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	fs, err := New(connStr)

	if err != nil {
		t.Error(err)
	}

	f, err := fs.TempFile("/dir1/dir2", "dir0")

	if err != nil {
		t.Error(err)
	}

	if f == nil {
		t.Error("File wasn't created")
	}

	dropTable(connStr)
}

func TestReadDir(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	fs, err := New(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"
	f, err := createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	dir := filepath.Dir(path)
	res, err := fs.ReadDir(dir)

	if err != nil {
		t.Error(err)
	}

	if res == nil {
		t.Error("No result")
	}

	fStat, err := f.Stat()

	if err != nil {
		t.Error(err)
	}

	if res[0].Name() != fStat.Name() {
		t.Errorf("Wrong file name. Must: %s, has: %s", fStat.Name(), res[0].Name())
	}

	if res[0].Size() != fStat.Size() {
		t.Errorf("Wrong file size. Must: %d, has: %d", fStat.Size(), res[0].Size())
	}

	dropTable(connStr)
}


func TestSymlink(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	fs, err := New(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"

	err = fs.Symlink("gdflhekj", path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	f, err := s.GetFile(path)

	if err != nil {
		t.Error(err)
	}

	if f == nil {
		t.Error("Symlink wasn't created")
	}

	dropTable(connStr)
}

func TestReadlink(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	fs, err := New(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"
	data := "gdflhekj"

	err = fs.Symlink(data, path)

	if err != nil {
		t.Error(err)
	}

	content, err := fs.Readlink(path)

	if err != nil {
		t.Error(err)
	}

	if data != content {
		t.Errorf("Wrong content. Must: %s, has: %s", data, content)
	}

	dropTable(connStr)
}

func TestGetFile(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"
	_, err = createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	f, err := s.GetFile(path)

	if err != nil {
		t.Error(err)
	}

	mustN := filepath.Base(path)
	hasN := f.Name()
	if mustN != hasN {
		t.Errorf("Wrong name! Must: %s, has: %s", mustN, hasN)
	}

	dropTable(connStr)
}

func TestGetFileID(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"
	f, err := createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	id, err := s.GetFileID(path)

	if err != nil {
		t.Error(err)
	}

	if f.ID != id {
		t.Errorf("Wrong id! Must: %d, has: %d", f.ID, id)
	}

	dropTable(connStr)
}

func TestRenameFile1(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path1 := "/dir1/dir2/file1.txt"
	f, err := createNewFile(path1)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	path2 := "/dir1/dir3/file1.txt"
	err = s.RenameFile(path1, path2)

	if err != nil {
		t.Error(err)
	}

	rF, err := s.GetFile(path2)

	if err != nil {
		t.Error(err)
	}

	if rF == nil {
		t.Error("File wasn't renamed")
	}

	if rF.Path != path2 {
		t.Errorf("Wrong path! Must: %s, has: %s", path2, rF.Path)
	}

	if f.ParentID == rF.ParentID {
		t.Error("ParentID wasn't changed!")
	}

	dropTable(connStr)
}

func TestRenameFile2(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path1 := "/dir1/dir2/file1.txt"
	f, err := createNewFile(path1)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	path2 := "/dir1/dir2/file2.txt"
	err = s.RenameFile(path1, path2)

	if err != nil {
		t.Error(err)
	}

	rF, err := s.GetFile(path2)

	if err != nil {
		t.Error(err)
	}

	if rF == nil {
		t.Error("File wasn't renamed")
	}

	if rF.Path != path2 {
		t.Errorf("Wrong path! Must: %s, has: %s", path2, rF.Path)
	}

	if f.ParentID != rF.ParentID {
		t.Error("ParentID was changed!")
	}

	dropTable(connStr)
}

func TestRenameFile3(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path1 := "/dir1/dir2"
	f, err := createNewFile(path1)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	path2 := "/dir3/dir2"
	err = s.RenameFile(path1, path2)

	if err != nil {
		t.Error(err)
	}

	rF, err := s.GetFile(path2)

	if err != nil {
		t.Error(err)
	}

	if rF == nil {
		t.Error("File wasn't renamed")
	}

	if rF.Path != path2 {
		t.Errorf("Wrong path! Must: %s, has: %s", path2, rF.Path)
	}

	if f.ParentID == rF.ParentID {
		t.Error("ParentID wasn't changed!")
	}

	dropTable(connStr)
}

func TestRenameFile4(t *testing.T) {
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path1 := "/dir1/dir2/file1.txt"
	_, err = createNewFile(path1)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	err = s.RenameFile("/dir1/dir2", "/dir1/dir3")

	if err != nil {
		t.Error(err)
	}

	f1, err := s.GetFile("/dir1/dir3/file1.txt")

	if err != nil {
		t.Error(err)
	}

	if f1 == nil {
		t.Error("Wrong path")
	}

	dropTable(connStr)
}
func TestRemoveFile1(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"
	_, err = createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	err = s.RemoveFile(path)

	if err != nil {
		t.Error(err)
	}

	getF, err := s.GetFile(path)

	if err != nil {
		t.Error(err)
	}

	if getF != nil {
		t.Error("File wasn't deleted")
	}

	dropTable(connStr)
}

func TestRemoveFile2(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2"
	_, err = createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	err = s.RemoveFile(path)

	if err != nil {
		t.Error(err)
	}

	getF, err := s.GetFile(path)

	if err != nil {
		t.Error(err)
	}

	if getF != nil {
		t.Error("File wasn't deleted")
	}
	
	dropTable(connStr)
}

func TestRemoveFile3(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"
	_, err = createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	dir := filepath.Dir(path)

	err = s.RemoveFile(dir)

	if err == nil {
		t.Errorf("No error! Must: `%s contains files`", dir)
	}

	if err.Error() != fmt.Sprintf("dir: %s contains files", dir) {
		t.Error(err)
	}

	dropTable(connStr)
}

func TestChildren1(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir0/dir1/dir2/file1.txt"
	_, err = createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	res, err := s.Children("/dir0")

	if err != nil {
		t.Error(err)
	}

	if len(res) != 1 {
		t.Errorf("Wrong children number. Must: %d, has: %d", 1, len(res))
	}

	dropTable(connStr)
}

func TestChildren2(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path1 := "/dir0/dir1/dir2/file1.txt"
	_, err = createNewFile(path1)

	if err != nil {
		t.Error(err)
	}

	path2 := "/dir0/dir4/dir2/file1.txt"
	_, err = createNewFile(path2)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	res, err := s.Children("/dir0")

	if err != nil {
		t.Error(err)
	}

	if len(res) != 2 {
		t.Errorf("Wrong children number. Must: %d, has: %d", 2, len(res))
	}
	
	dropTable(connStr)
}

func TestChildrenByFileID(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir0/dir1/dir2/file1.txt"
	_, err = createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	path2 := "/dir0"
	dir, err := s.GetFile(path2)

	if err != nil {
		t.Error(err)
	}

	if dir == nil {
		t.Errorf("No such file path: %s", path2)
	}

	res, err := s.ChildrenByFileID(dir.ID)

	if err != nil {
		t.Error(err)
	}

	if len(res) != 1 {
		t.Errorf("Wrong children number. Must: %d, has: %d", 1, len(res))
	}

	dropTable(connStr)
}

func TestChildrenIdsByFileID(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir0/dir1/dir2/file1.txt"
	_, err = createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	path2 := "/dir0"
	dir, err := s.GetFile(path2)

	if err != nil {
		t.Error(err)
	}

	if dir == nil {
		t.Errorf("No such file path: %s", path2)
	}

	res, err := s.ChildrenIdsByFileID(dir.ID)

	if err != nil {
		t.Error(err)
	}

	path3 := "/dir0/dir1"
	f, err := s.GetFile(path3)

	if err != nil {
		t.Error(err)
	}

	if f == nil {
		t.Errorf("No such file path: %s", path2)
	}

	l := len(res)

	if l != 1 {
		t.Errorf("Wrong children number. Must: %d, has: %d", 1, l)
	}

	if f.ID != res[0] {
		t.Errorf("Wrong children number. Must: %d, has: %d", 1, len(res))
	}

	dropTable(connStr)
}

func TestUpdateFileContent(t *testing.T){
	err := createTable(connStr)

	if err != nil {
		t.Error(err)
	}

	path := "/dir1/dir2/file1.txt"
	f, err := createNewFile(path)

	if err != nil {
		t.Error(err)
	}

	s, err := newStorage(connStr)

	if err != nil {
		t.Error(err)
	}

	c := []byte("11111")

	err = s.UpdateFileContent(f.ID, c)

	if err != nil {
		t.Error(err)
	}

	f1, err := s.GetFile(path)

	if err != nil {
		t.Error(err)
	}

	if bytes.Equal(f.Content, f1.Content) {
		t.Error("File content wasn't changed")
	}

	c2 := []byte("222222")

	err = s.UpdateFileContent(f.ID, c2)

	if err != nil {
		t.Error(err)
	}

	f2, err := s.GetFile(path)

	if err != nil {
		t.Error(err)
	}

	if bytes.Equal(f.Content, f2.Content) {
		t.Error("File content wasn't changed")
	}

	dropTable(connStr)
}

func createNewFile(path string) (*File, error) {
	s, err := newStorage(connStr)

	if err != nil {
		return nil, err
	}

	f, err := s.NewFile(path, 0666, os.O_RDWR|os.O_CREATE|os.O_TRUNC)

	if err != nil {
		return nil, err
	}

	return f, nil
}

func connectToDB(connStr string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return db, nil
}

func createTable(connStr string) error {
	db, err := sqlx.Connect("mysql", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	db.MustExec(schema)

	return nil
}

func dropTable(connStr string) error {
	db, err := sqlx.Connect("mysql", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	db.MustExec("DROP TABLE `file`")

	return nil
}
