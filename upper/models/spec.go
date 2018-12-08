package models

// Spec .
type Spec struct {
	ID        string `db:"id"`
	Weight    int    `db:"weight"`
	ProductID string `db:"product_id"`
}
