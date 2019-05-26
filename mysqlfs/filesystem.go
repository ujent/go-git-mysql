// Package mysqlfs provides a billy filesystem base on mysql db
package mysqlfs

import (
	"errors"
	"io"
	"os"
	"time"

	"gopkg.in/src-d/go-billy.v4"
	"gopkg.in/src-d/go-billy.v4/helper/chroot"
)

//Mysqlfs - realization of billy.Flesystem based on MySQL
type Mysqlfs struct {
	storage *storage
}

//New creates an instance of billy.Filesystem
func New(connectionStr string) (billy.Filesystem, error) {
	storage, err := newStorage(connectionStr)

	if err != nil {
		return nil, err
	}

	fs := &Mysqlfs{storage: storage}

	return chroot.New(fs, string(separator)), nil
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func (*Mysqlfs) Create(filename string) (billy.File, error) {
	return nil, nil
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (*Mysqlfs) Open(filename string) (billy.File, error) {
	return nil, nil
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.

func (*Mysqlfs) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	return nil, nil
}

// Stat returns a FileInfo describing the named file.
func (*Mysqlfs) Stat(filename string) (os.FileInfo, error) {
	return nil, nil
}

// Rename renames (moves) oldpath to newpath. If newpath already exists and
// is not a directory, Rename replaces it. OS-specific restrictions may
// apply when oldpath and newpath are in different directories.
func (*Mysqlfs) Rename(oldpath, newpath string) error {
	return nil
}

// Remove removes the named file or directory.
func (*Mysqlfs) Remove(filename string) error {
	return nil
}

// Join joins any number of path elements into a single path, adding a
// Separator if necessary. Join calls filepath.Clean on the result; in
// particular, all empty strings are ignored. On Windows, the result is a
// UNC path if and only if the first path element is a UNC path.
func (*Mysqlfs) Join(elem ...string) string {
	return ""
}

// TempFile creates a new temporary file in the directory dir with a name
// beginning with prefix, opens the file for reading and writing, and
// returns the resulting *os.File. If dir is the empty string, TempFile
// uses the default directory for temporary files (see os.TempDir).
// Multiple programs calling TempFile simultaneously will not choose the
// same file. The caller can use f.Name() to find the pathname of the file.
// It is the caller's responsibility to remove the file when no longer
// needed.
func (*Mysqlfs) TempFile(dir, prefix string) (billy.File, error) {
	return nil, nil
}

// ReadDir reads the directory named by dirname and returns a list of
// directory entries sorted by filename.
func (*Mysqlfs) ReadDir(path string) ([]os.FileInfo, error) {
	return nil, nil
}

// MkdirAll creates a directory named path, along with any necessary
// parents, and returns nil, or else returns an error. The permission bits
// perm are used for all directories that MkdirAll creates. If path is/
// already a directory, MkdirAll does nothing and returns nil.
func MkdirAll(filename string, perm os.FileMode) error {
	return nil
}

// Lstat returns a FileInfo describing the named file. If the file is a
// symbolic link, the returned FileInfo describes the symbolic link. Lstat
// makes no attempt to follow the link.
func (*Mysqlfs) Lstat(filename string) (os.FileInfo, error) {
	return nil, nil
}

// Symlink creates a symbolic-link from link to target. target may be an
// absolute or relative path, and need not refer to an existing node.
// Parent directories of link are created as necessary.
func (*Mysqlfs) Symlink(target, link string) error {
	return nil
}

// Readlink returns the target path of link.
func (*Mysqlfs) Readlink(link string) (string, error) {
	return "", nil
}

// Chroot returns a new filesystem from the same type where the new root is
// the given path. Files outside of the designated directory tree cannot be
// accessed.
func (*Mysqlfs) Chroot(path string) (billy.Filesystem, error) {
	return nil, nil
}

// Root returns the root path of the filesystem.
func (*Mysqlfs) Root() string {
	return ""
}

// Capabilities implements the Capable interface.
func (fs *Mysqlfs) Capabilities() billy.Capability {
	return billy.WriteCapability |
		billy.ReadCapability |
		billy.ReadAndWriteCapability |
		billy.SeekCapability |
		billy.TruncateCapability
}

func (f *File) Name() string {
	return f.FileName
}

func (f *File) Read(b []byte) (int, error) {
	n, err := f.ReadAt(b, f.Position)
	f.Position += int64(n)

	if err == io.EOF && n != 0 {
		err = nil
	}

	return n, err
}

func (f *File) ReadAt(b []byte, off int64) (int, error) {
	if f.IsClosed {
		return 0, os.ErrClosed
	}

	if !isReadAndWrite(f.Flag) && !isReadOnly(f.Flag) {
		return 0, errors.New("read not supported")
	}

	n, err := readAt(f.Content, b, off)

	return n, err
}

func readAt(content []byte, b []byte, off int64) (n int, err error) {
	size := int64(len(content))
	if off >= size {
		return 0, io.EOF
	}

	l := int64(len(b))
	if off+l > size {
		l = size - off
	}

	btr := content[off : off+l]
	if len(btr) < len(b) {
		err = io.EOF
	}
	n = copy(b, btr)

	return
}

func writeAt(content []byte, p []byte, off int64) (int, error) {
	prev := len(content)

	diff := int(off) - prev
	if diff > 0 {
		content = append(content, make([]byte, diff)...)
	}

	content = append(content[:off], p...)
	if len(content) < prev {
		content = content[:prev]
	}

	return len(p), nil
}

func saveFileToDb() error {
	return nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.IsClosed {
		return 0, os.ErrClosed
	}

	switch whence {
	case io.SeekCurrent:
		f.Position += offset
	case io.SeekStart:
		f.Position = offset
	case io.SeekEnd:
		f.Position = int64(len(f.Content)) + offset
	}

	return f.Position, nil
}

func (f *File) Write(p []byte) (int, error) {
	if f.IsClosed {
		return 0, os.ErrClosed
	}

	if !isReadAndWrite(f.Flag) && !isWriteOnly(f.Flag) {
		return 0, errors.New("write not supported")
	}

	n, err := writeAt(f.Content, p, f.Position)
	f.Position += int64(n)

	return n, err
}

func (f *File) Close() error {
	if f.IsClosed {
		return os.ErrClosed
	}

	f.IsClosed = true
	return nil
}

func (f *File) Truncate(size int64) error {
	if size < int64(len(f.Content)) {
		f.Content = f.Content[:size]
	} else if more := int(size) - len(f.Content); more > 0 {
		f.Content = append(f.Content, make([]byte, more)...)
	}

	return nil
}

func (f *File) Duplicate(filename string, mode os.FileMode, flag int) billy.File {
	new := &File{
		FileName: filename,
		Content:  f.Content,
		Mode:     mode,
		Flag:     flag,
	}

	if isAppend(flag) {
		new.Position = int64(len(new.Content))
	}

	if isTruncate(flag) {
		new.Content = make([]byte, 0)
	}

	return new
}

func (f *File) Stat() (os.FileInfo, error) {
	return &FileInfo{
		FileName: f.Name(),
		FileMode: f.Mode,
		FileSize: int64(len(f.Content)),
	}, nil
}

// Lock is a no-op in memfs.
func (f *File) Lock() error {
	return nil
}

// Unlock is a no-op in memfs.
func (f *File) Unlock() error {
	return nil
}

func (fi *FileInfo) Name() string {
	return fi.FileName
}

func (fi *FileInfo) Size() int64 {
	return int64(fi.FileSize)
}

func (fi *FileInfo) Mode() os.FileMode {
	return fi.FileMode
}

func (*FileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *FileInfo) IsDir() bool {
	return fi.FileMode.IsDir()
}

func (*FileInfo) Sys() interface{} {
	return nil
}

func isCreate(flag int) bool {
	return flag&os.O_CREATE != 0
}

func isAppend(flag int) bool {
	return flag&os.O_APPEND != 0
}

func isTruncate(flag int) bool {
	return flag&os.O_TRUNC != 0
}

func isReadAndWrite(flag int) bool {
	return flag&os.O_RDWR != 0
}

func isReadOnly(flag int) bool {
	return flag == os.O_RDONLY
}

func isWriteOnly(flag int) bool {
	return flag&os.O_WRONLY != 0
}

func isSymlink(m os.FileMode) bool {
	return m&os.ModeSymlink != 0
}
