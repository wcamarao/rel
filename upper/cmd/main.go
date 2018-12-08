package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wcamarao/rel/upper/models"
	db "upper.io/db.v3"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/postgresql"
)

var settings = postgresql.ConnectionURL{
	Database: "rel",
}

func main() {
	sess, err := postgresql.Open(settings)
	if err != nil {
		log.Fatalf("Open: %v", err)
	}
	// sess.SetLogging(true)
	defer sess.Close()

	// Settings

	log.Printf(
		"Settings: ConnMaxLifetime: %d, MaxIdleConns: %d, MaxOpenConns: %d",
		sess.ConnMaxLifetime(),
		sess.MaxIdleConns(),
		sess.MaxOpenConns(),
	)

	// Truncate

	collections := []string{"category", "image", "product", "product_category", "spec"}
	for _, collection := range collections {
		if err = sess.Collection(collection).Truncate(); err != nil {
			log.Fatalf("Truncate: %v", err)
		}
	}

	// Create Products

	createProduct("foo", "Foo", sess.InsertInto("product"))
	createProduct("bar", "Barr", sess.InsertInto("product"))
	findProducts(sess.Collection("product"))

	// Transaction

	sess.Tx(context.Background(), func(tx sqlbuilder.Tx) error {
		createProduct("zip", "Zip", tx.InsertInto("product"))

		_, err = tx.Update("product").
			Set("name", "Bar").
			Where("id", "bar").
			Exec()
		if err != nil {
			log.Fatalf("Update: %v", err)
			return err
		}

		return nil
	})

	findProducts(sess.Collection("product"))

	// Create Specs

	createSpec("fspec", 1, "foo", sess.InsertInto("spec"))
	createSpec("bspec", 2, "bar", sess.InsertInto("spec"))
	createSpec("zspec", 3, "zip", sess.InsertInto("spec"))
	findSpecs(sess.Collection("spec"))
}

func createProduct(id, name string, inserter sqlbuilder.Inserter) {
	now := time.Now()
	product := models.Product{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := inserter.Values(product).Exec()
	if err != nil {
		log.Fatalf("InsertInto: %v", err)
	}
}

func findProducts(collection db.Collection) {
	var products []models.Product
	if err := collection.Find().OrderBy("created_at").All(&products); err != nil {
		log.Fatalf("Find: %v", err)
	}

	values := []string{}
	for _, product := range products {
		values = append(values, fmt.Sprintf("%s:%s", product.ID, product.Name))
	}
	log.Printf("Products: %v", values)
}

func createSpec(id string, weight int, productId string, inserter sqlbuilder.Inserter) {
	spec := models.Spec{
		ID:        id,
		Weight:    weight,
		ProductID: productId,
	}

	_, err := inserter.Values(spec).Exec()
	if err != nil {
		log.Fatalf("InsertInto: %v", err)
	}
}

func findSpecs(collection db.Collection) {
	var specs []models.Spec
	if err := collection.Find().OrderBy("weight").All(&specs); err != nil {
		log.Fatalf("Find: %v", err)
	}

	values := []string{}
	for _, spec := range specs {
		values = append(values, fmt.Sprintf("%s:%d", spec.ID, spec.Weight))
	}
	log.Printf("Products: %v", values)
}
