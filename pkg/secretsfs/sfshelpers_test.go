package secretsfs

import (
	"testing"
)

func TestRootName(t *testing.T) {
	// t.Fatal("not implemented")
	tables := []struct {
		npath    string
		rootpath string
		subpath  string
	}{
		{"/secretsfs/mydir/myfile.txt", "secretsfs", "mydir/myfile.txt"},
		{"/", "", ""},
		{"/tests", "tests", ""},
		{"/secretsfs/myfile.txt", "secretsfs", "myfile.txt"},
	}

	for _, table := range tables {
		rootpath, subpath := rootName(table.npath)
		if rootpath != table.rootpath {
			t.Errorf("rootpath of '%v' was incorrect, got: '%v', want: '%v'\n", table.npath, rootpath, table.rootpath)
		}
		if subpath != table.subpath {
			t.Errorf("subpath of '%v' was incorrect, got: '%v', want: '%v'\n", table.npath, subpath, table.subpath)
		}
	}
}
