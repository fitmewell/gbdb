package bdb

import (
	"reflect"
	"sync"
	"strconv"
	"errors"
	"strings"
)

type BColumn struct {
	SqlName      string
	Field        reflect.StructField
	Index        int
	IsPrimaryKey bool
}

// only used as a place holder for tag definition for table name
type BTableName struct {
}

var tableNameType = reflect.TypeOf(BTableName{})

type BTable struct {
	SqlName    string
	Type       reflect.Type
	PrimaryKey BColumn
	Columns    []BColumn
	ColumnMap  map[string]BColumn
}

// define to store reflect in memory
var cachedDefinition = map[reflect.Type]BTable{}

// define for map definition
var cacheMutex = sync.Mutex{}

func GetDefinition(rType reflect.Type) (bTable BTable, err error) {
	bTable, ok := cachedDefinition[rType]
	if ok {
		return bTable, err
	}
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	bTable, ok = cachedDefinition[rType]
	if ok {
		return bTable, err
	}
	bTable, err = genDefinition(rType)
	if err != nil {
		return bTable, err
	}
	cachedDefinition[rType] = bTable
	return bTable, err
}

const sqlName = "name"
const sqlIndex = "index"
const sqlPrimaryKey = "primaryKey"

// gen table sql definition from struct reflect info
func genDefinition(t reflect.Type) (bTable BTable, err error) {

	fields := getFields(t)

	cacheMap := map[string]int{}

	foundPrimary := false
	for _, field := range fields {
		if _, ok := cacheMap[field.Name]; ok {
			continue
		}
		if field.Type == tableNameType {
			bTable.SqlName = field.Tag.Get(sqlName)
		} else {
			column := BColumn{Field: field}
			if field.Tag.Get(sqlIndex) != "" {
				column.Index, err = strconv.Atoi(field.Tag.Get(sqlIndex))
				if err != nil {
					return bTable, err
				}
			}
			if field.Tag.Get(sqlName) != "" {
				column.SqlName = field.Tag.Get(sqlName)
			}
			if field.Tag.Get(sqlPrimaryKey) != "" {
				column.IsPrimaryKey, err = strconv.ParseBool(field.Tag.Get(sqlPrimaryKey))
				if column.IsPrimaryKey {
					if !foundPrimary {
						bTable.PrimaryKey = column
						foundPrimary = true
					} else {
						return bTable, errors.New("duplicate primary key found")
					}
				}
			}
			cacheMap[field.Name] = 1
			bTable.Columns = append(bTable.Columns, column)
		}
	}

	for _, column := range bTable.Columns {
		if _, ok := bTable.ColumnMap[column.SqlName]; ok {
			continue
		}
		bTable.ColumnMap[column.SqlName] = column
	}

	if bTable.SqlName == "" {
		bTable.SqlName = humpNamed(t.Name())
	}

	return bTable, err
}

const uppers = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func humpNamed(source string) string {
	resp := ""
	for i := 0; i < len(source); i++ {
		s := string(source[i])
		if strings.IndexAny(uppers, s) != -1 {
			resp += "_" + s
		} else {
			resp += s
		}
	}
	return resp
}

func getFields(t reflect.Type) []reflect.StructField {
	var fields []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			fields = append(fields, getFields(f.Type)...)
		}
		fields = append(fields, f)
	}
	return fields
}
