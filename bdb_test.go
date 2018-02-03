package bdb

import (
	"testing"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
)

type helloWorld struct {
	Id   string `autoIncreased:"true"`
	Name string
}

func Test_gbdb_Insert(t *testing.T) {

	db, err := sql.Open("mysql", "root:1991@tcp(localhost:3306)/test")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	g := gbdb{db: db}

	world := helloWorld{Name: "single"}
	result, err := g.Insert(world, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)

	worlds := []helloWorld{
		{Name: "multi"},
		{Name: "multi"},
		{Name: "multi"},
		{Name: "multi"},
	}

	result, err = g.Insert(worlds, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)
}
