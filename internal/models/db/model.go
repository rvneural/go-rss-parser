package db

type RSS struct {
	ID    int64  `db:"id"`
	URL   string `db:"url"`
	Title string `db:"title"`
}
