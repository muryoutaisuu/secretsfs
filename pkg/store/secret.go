package store

type Secret struct {
	Path    string
	Mode    int
	Content string
	Subs    []*Secret
}
