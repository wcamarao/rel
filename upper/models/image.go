package models

// Image .
type Image struct {
	ID        string `db:"id"`
	URL       string `db:"url"`
	ProductID string `db:"product_id"`
}
