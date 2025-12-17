package changelogs

type ChangeItem struct {
	Text string
}

type ChangeSection struct {
	Title string
	Items []ChangeItem
}

type ChangeLog struct {
	Version string
	Date    string
	Anchor  string
	Sections []ChangeSection
}

