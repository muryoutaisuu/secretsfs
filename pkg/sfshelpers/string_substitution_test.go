package sfshelpers

import (
	"testing"
	"reflect"
)

func TestSubstitutionPossibilities(t *testing.T) {
	tables := []struct {
		s,b,n string
		z []string
	}{
		{"a_b_c_d", "_", "/", []string {"a_b_c_d", "a_b_c/d", "a_b/c_d", "a_b/c/d", "a/b_c_d", "a/b_c/d", "a/b/c_d", "a/b/c/d"}},
		{"_c", "_", "/", []string {"_c", "/c"}},
		{"_c_", "_", "/", []string {"_c_", "_c/", "/c_", "/c/"}},
	}

	for _, table := range tables {
		value := SubstitutionPossibilities(table.s, table.b, table.n)
		if !reflect.DeepEqual(value, table.z) {
			t.Errorf("SubstitutionPossibilities(%s, %s, %s) was incorrect, got %v, want: %v", table.s, table.b, table.n, value, table.z)
		}
	}
}


func TestInv(t *testing.T) {
	tables := []struct {
		s,b,n,z string
	}{
		{"_", "_", "/", "/"},
		{"foo", "_", "/", "f"},
		{"_foo", "_", "/", "/"},
		{"/foo", "_", "/", "/"},
		{"/foo", "/", "_", "_"},
		{"", "_", "/", ""},
	}

	for _, table := range tables {
		value := inv(table.s, table.b, table.n)
		if value != table.z {
			t.Errorf("inv(%s, %s, %s) was incorrect, got %s, want: %s", table.s, table.b, table.n, value, table.z)
		}
	}
}

func TestNeedInv(t *testing.T) {
	tables := []struct {
		s,b string
		z bool
	}{
		{"_yes", "_", true},
		{"no", "_", false},
		{"", "_", false},
	}

	for _, table := range tables {
		need := needInv(table.s, table.b)
		if need != table.z {
			t.Errorf("needInv(%s, %s) was incorrect, got %t, want: %t", table.s, table.b, need, table.z)
		}
	}
}
