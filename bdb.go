package bdb

import (
	"database/sql"
	"errors"
	"reflect"
)

/*
TODO 1 add map support
*/



type gbdb struct {
	db *sql.DB
}

// Insert for auto insert
// vs the interface to be inserted must be struct or list or slice or map
// onlyNoneNil true if ignore nil value , only available when vs is struct
func (g *gbdb) Insert(v interface{}, onlyNoneNil bool) (sql.Result, error) {

	vt := reflect.TypeOf(v)
	vv := reflect.ValueOf(v)

	for vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
		vv = vv.Elem()
	}

	isCollections := false

	if vt.Kind() == reflect.Array || vt.Kind() == reflect.Slice {
		if vv.Len() == 0 {
			return nil, errors.New("empty array is not allowed here ")
		}
		vt = vt.Elem()
		for vt.Kind() == reflect.Ptr {
			vt = vt.Elem()
		}
		isCollections = true
	}

	bTable, err := GetDefinition(vt)
	if err != nil {
		return nil, err
	}

	cNumber := len(bTable.Columns)
	if cNumber == 0 {
		return nil, errors.New("no  columns defined in this struct")
	}

	query := "INSERT INTO `" + bTable.SQLName + "` ("
	propPlaceholder := ""
	var props []interface{}

	for i, column := range bTable.Columns {
		if column.IsAutoIncreased {
			continue
		}
		if onlyNoneNil && !isCollections {
			value := column.GetValue(vt)
			if value == "" {
				continue
			}
		}
		query += "`" + column.SQLName + "`"
		propPlaceholder += "?"
		if i != cNumber-1 {
			query += ","
		}
	}

	query += ") VALUES "

	var vls []interface{}
	if isCollections {
		vl := vv.Len()
		for i := 0; i < vl; i++ {
			index := vv.Index(i)
			for index.Kind() == reflect.Ptr {
				index = index.Elem()
			}
			vls = append(vls, index.Interface())
		}
	} else {
		vls = []interface{}{vv.Interface()}
	}

	for i, v := range vls {
		query += "("
		for ci, column := range bTable.Columns {
			if column.IsAutoIncreased {
				continue
			}
			query += "?"
			props = append(props, column.GetValue(v))
			if ci != cNumber-1 {
				query += ","
			}
		}
		query += ")"
		if i != len(vls)-1 {
			query += ","
		}
	}

	stmt, err := g.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(props...)
	return result, err
}