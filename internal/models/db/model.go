package db

type RSS struct {
	ID    int64  `db:"id"`
	URL   string `db:"url" json:"url"`
	Title string `db:"title" json:"title"`
}
