package store

type Secret struct {
	Path    string
	Mode    int64
	Content string
	Subs    []*Secret
}
