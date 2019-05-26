package mysqlfs

import "testing"

func TestAddTable(t *testing.T) {
	_, err := newStorage("user:password@/dbname")

	if err != nil {
		panic(err)
	}
}
