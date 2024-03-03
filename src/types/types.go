package types

type Args struct {
	Command  string
	Input    string
	Name     string
	Type     string
	Target   string
	Sequence int
	Verbose  bool
	Pwd      string
}

type Replacement struct {
	Type    string
	Value   string
	Options []string
	Indents int
	Target  string
}

type Page struct {
	Filename string
	UrlName  string
	Name     string
	Path     string
	Type     string
	Content  string
	Sequence int
}

type Pagegroup struct {
	Ident   string
	Entries []Page
}
