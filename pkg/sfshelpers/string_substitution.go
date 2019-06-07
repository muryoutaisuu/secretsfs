package sfshelpers

// SubstitutionPossibilities calculates all different possibilities that can be
// be achieved by substituting a string one way or another.
// Takes string as arg s, then a bad string as b, and a new string as n.
// Also look at this example:
//
//	func main() {
//		orig := "a_b_c_d"
//		mytest := SubstitutionPossibilities(orig, "_", "/")
//		fmt.Printf("orig: %s\nvariants: %s", orig, mytest)
//	}
//	
// output will be:
//	orig: a_b_c_d
//	variants: [a_b_c_d a_b_c/d a_b/c_d a_b/c/d a/b_c_d a/b_c/d a/b/c_d a/b/c/d]
func SubstitutionPossibilities(s, b, n string) []string {
	l := len(s)
	if l == 0 {
		return []string{}
	}
	var mys []string

	if l > 1 {
		mys = SubstitutionPossibilities(s[1:l], b, n)
	} else {
		mys = []string{s}
		if needInv(s, b) {
			mys = append(mys, inv(s,b,n))
		}
		return mys
	}

	lm := len(mys)
	for i:=0; i<lm; i++ {
		mys[i] = string(s[0])+mys[i]
		if needInv(s,b) {
			mys = append(mys, inv(s,b,n,)+mys[i][1:])
		}
	}
	return mys
}

func inv(s, b, n string) string {
	if len(s) < 1 {
		return ""
	}

	c := string(s[0])
	if c == b {
		return n
	}
	return c
}

func needInv(s, b string) bool {
	if len(s) < 1 {
		return false
	}

	if string(s[0]) == b {
		return true
	}
	return false
}
