package secretsfs

import (
	"testing"
)

// getDirEntries returns all direntries matching spath in entries
// npath:   "/tests"
// entries: "/tests/dir1"
//          "/tests/file1"
//          "/tests/dir1/file2"
//          "/otherdir"
// returns: "/tests/dir1"
//          "/tests/file1"
// not tested against edge cases, like entry "/testsdir1"
// only used for fio_tests.go, do *NOT* use in your projects
func TestGetDirEntries(t *testing.T) {
	// t.Fatal("not implemented")
	var testnodes = testNodes{
		[]*testNode{
			&testNode{"/tests", false, nil},
			&testNode{"/tests/testdir", false, nil},
			&testNode{"/tests/test1.txt", true, []byte("This is the content of the file /tests/test1.txt")},
			&testNode{"/tests/test2.txt", true, []byte("This is the content of the file /tests/test2.txt")},
			&testNode{"/tests/testdir/test3.txt", true, []byte("This is the content of the file /tests/testdir/test3.txt")},
			&testNode{"/tests/testdir/test4.txt", true, []byte("This is the content of the file /tests/testdir/test4.txt")},
			&testNode{"/tests/testdir/test5.txt", true, []byte("This is the content of the file /tests/testdir/test5.txt")},
		},
	}
	tables := []struct {
		npath   string
		returns []string
	}{
		{"/tests", []string{"/tests/testdir", "/tests/test1.txt", "/tests/test2.txt"}},
		{"/tests/testdir", []string{"/tests/testdir/test3.txt", "/tests/testdir/test4.txt", "/tests/testdir/test5.txt"}},
	}
	for _, table := range tables {
		var entrylist []string
		for _, direntry := range testnodes.getDirEntries(table.npath) {
			entrylist = append(entrylist, direntry.path)
		}
		if len(entrylist) != len(table.returns) {
			t.Errorf("Not the corrent amount of entries returned for npath='%v'!\nGot:  %v\nWant: %v\n", table.npath, entrylist, table.returns)
			break
		}
		for k := range entrylist {
			if entrylist[k] != table.returns[k] {
				t.Errorf("Path returned was incorrect for npath='%v'!\nGot:  %v\nWant: %v\n", table.npath, entrylist[k], table.returns[k])
			}
		}
	}
	//for _, table := range tables {
	//	e := getDirEntries(
	//	rootpath, subpath := rootName(table.npath)
	//	if rootpath != table.rootpath {
	//		t.Errorf("rootpath of '%v' was incorrect, got: '%v', want: '%v'\n", table.npath, rootpath, table.rootpath)
	//	}
	//	if subpath != table.subpath {
	//		t.Errorf("subpath of '%v' was incorrect, got: '%v', want: '%v'\n", table.npath, subpath, table.subpath)
	//	}
	//}
}
