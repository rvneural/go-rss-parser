package db

type RSS struct {
	ID    int64  `db:"id" json:"id"`
	URL   string `db:"url" json:"url"`
	Title string `db:"title" json:"title"`
	Type  string `db:"type" json:"type"`
}
