package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jmoiron/modl"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/wcamarao/rel/sqlx/models"
)

func main() {
	db, err := sqlx.Connect("postgres", "dbname=rel sslmode=disable")
	if err != nil {
		log.Fatalf("Open: %v", err)
	}
	defer db.Close()

	db.MapperFunc(strcase.ToSnake)
	modl.TableNameMapper = strcase.ToSnake
	sqlx.NameMapper = strcase.ToSnake
	dbmap := modl.NewDbMap(db.DB, modl.PostgresDialect{})
	dbmap.TraceOn("[SQL]", log.New(os.Stdout, "", log.LUTC))
	dbmap.AddTable(models.Category{})
	dbmap.AddTable(models.Image{})
	dbmap.AddTable(models.Product{}).SetKeys(false, "id")
	dbmap.AddTable(models.Spec{})
	fmt.Printf("%+v\n", db.Stats())

	// Truncate

	dbmap.TruncateTables()

	// Create Products

	createProduct("foo", "Foo", dbmap)
	createProduct("bar", "Barr", dbmap)
	findProducts(dbmap)

	// Transaction

	tx, err := dbmap.Begin()
	if err != nil {
		log.Fatalf("Begin: %v", err)
	}

	createProduct("zip", "Zip", tx)

	bar := &models.Product{}
	if err = tx.Get(bar, "bar"); err != nil {
		log.Fatalf("Get: %v", err)
	}

	bar.Name = "Bar"
	if _, err = tx.Update(bar); err != nil {
		log.Fatalf("Update: %v", err)
	}

	if err = tx.Commit(); err != nil {
		log.Fatalf("Commit: %v", err)
	}

	findProducts(dbmap)

	// Create Specs

	createSpec("fspec", 1, "foo", dbmap)
	createSpec("bspec", 2, "bar", dbmap)
	createSpec("zspec", 3, "zip", dbmap)
	findSpecs(dbmap)

	// Named Query

	rows, err := db.NamedQuery("select * from spec where weight > :weight order by weight", map[string]interface{}{"weight": 1})
	if err != nil {
		log.Fatalf("NamedQuery: %v", err)
	}

	spec := models.Spec{}
	for rows.Next() {
		err = rows.StructScan(&spec)
		if err != nil {
			log.Fatalf("StructScan: %v", err)
		}
		log.Printf("Spec: %v", spec)
	}

	// Join 1-1

	type ProductSpec struct {
		models.Product `db:"p"`
		models.Spec    `db:"s"`
	}

	jFields := joinFields(map[string]interface{}{
		"p": &models.Product{},
		"s": &models.Spec{},
	})

	productSpecs := []ProductSpec{}
	q := fmt.Sprintf(`SELECT %s FROM product p JOIN spec s ON p.id = s.product_id`, jFields)
	err = dbmap.Select(&productSpecs, q)
	if err != nil {
		log.Fatalf("Join 1-1: %v", err)
	}
	for _, ps := range productSpecs {
		fmt.Println(ps.Product, "--", ps.Spec)
	}

	// Create Images

	createImage("fgif", "foo.gif", "foo", dbmap)
	createImage("fpng", "foo.png", "foo", dbmap)
	findImages(dbmap)

	// Join 1-N

	type ProductImages struct {
		models.Product `db:"p"`
		models.Image   `db:"i"`
	}

	jFields = joinFields(map[string]interface{}{
		"p": &models.Product{},
		"i": &models.Image{},
	})

	productImages := []ProductImages{}
	q = fmt.Sprintf(`SELECT %s FROM product p JOIN image i ON p.id = i.product_id`, jFields)
	err = dbmap.Select(&productImages, q)
	if err != nil {
		log.Fatalf("Join 1-N: %v", err)
	}
	for _, pi := range productImages {
		fmt.Println(pi.Product, "--", pi.Image)
	}
}

func joinFields(types map[string]interface{}) string {
	fields := []string{}
	for a, typ := range types {
		el := reflect.ValueOf(typ).Elem()
		for i := 0; i < el.NumField(); i++ {
			f := strcase.ToSnake(el.Type().Field(i).Name)
			fields = append(fields, fmt.Sprintf(`%s.%s "%s.%s"`, a, f, a, f))
		}
	}
	return strings.Join(fields, ", ")
}

func createProduct(id, name string, se modl.SqlExecutor) {
	now := time.Now()
	err := se.Insert(&models.Product{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		log.Fatalf("Insert: %v", err)
	}
}

func findProducts(se modl.SqlExecutor) {
	products := []models.Product{}
	err := se.Select(&products, "select * from product order by created_at")
	if err != nil {
		log.Fatalf("Select: %v", err)
	}

	values := []string{}
	for _, product := range products {
		values = append(values, fmt.Sprintf("%s:%s", product.ID, product.Name))
	}
	log.Printf("Products: %v", values)
}

func createSpec(id string, weight int, productId string, se modl.SqlExecutor) {
	err := se.Insert(&models.Spec{
		ID:        id,
		Weight:    weight,
		ProductID: productId,
	})
	if err != nil {
		log.Fatalf("Insert: %v", err)
	}
}

func findSpecs(se modl.SqlExecutor) {
	specs := []models.Spec{}
	err := se.Select(&specs, "select * from spec order by weight")
	if err != nil {
		log.Fatalf("Select: %v", err)
	}

	values := []string{}
	for _, spec := range specs {
		values = append(values, fmt.Sprintf("%s:%d", spec.ID, spec.Weight))
	}
	log.Printf("Specs: %v", values)
}

func createImage(id, url, productId string, se modl.SqlExecutor) {
	err := se.Insert(&models.Image{
		ID:        id,
		URL:       url,
		ProductID: productId,
	})
	if err != nil {
		log.Fatalf("Insert: %v", err)
	}
}

func findImages(se modl.SqlExecutor) {
	images := []models.Image{}
	err := se.Select(&images, "select * from image order by url")
	if err != nil {
		log.Fatalf("Select: %v", err)
	}

	values := []string{}
	for _, image := range images {
		values = append(values, fmt.Sprintf("%s:%s", image.ID, image.URL))
	}
	log.Printf("Images: %v", values)
}
