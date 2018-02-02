package bdb

import (
	"database/sql"
	"errors"
	"reflect"
)

type gbdb struct {
	db *sql.DB
}

func (g *gbdb) Insert(v interface{}) (sql.Result, error) {

	bTable, err := GetDefinition(reflect.TypeOf(v))
	if err != nil {
		return nil, err
	}

	cNumber := len(bTable.Columns)
	if cNumber == 0 {
		return nil, errors.New("no  columns defined in this struct")
	}

	query := "INSERT INTO `" + bTable.SQLName + "` ("
	propPlaceholder := ""
	props := []interface{}{}

	for i, column := range bTable.Columns {
		query += "'" + column.SQLName + "'"
		propPlaceholder += "?"
		if i != cNumber-1 {
			query += ","
			propPlaceholder += ","
		}
		props = append(props, column.GetValue(v))
	}

	query += ") VALUES (" + propPlaceholder + ")"

	stmt, err := g.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(props...)
	return result, err
}

func (g *gbdb) InsertSelective() (sql.Result, error) {
	stmt, err := g.db.Prepare("INSERT INTO `` VALUES ()")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	result, e := stmt.Exec("")
	return result, e
}
