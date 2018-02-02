package bdb

import (
	"errors"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// BColumn all column generated from struct
type BColumn struct {
	SQLName      string
	Field        reflect.StructField
	FieldIndex   int
	Index        int
	IsPrimaryKey bool
}

// GetValue get selected value
func (bc *BColumn) GetValue(v interface{}) string {
	vf := reflect.ValueOf(v)
	for vf.Kind() == reflect.Ptr {
		vf = vf.Elem()
	}
	fv := vf.Field(bc.Index)
	for fv.Kind() == reflect.Ptr {
		fv = fv.Elem()
	}
	ft := bc.Field.Type

	str := ""
	switch ft.Kind() {
	case reflect.Invalid:
		log.Println("unmatched type found")
	case reflect.Bool:
		if fv.Bool() {
			str = "1"
		} else {
			str = "0"
		}
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		str = string(fv.Int())
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		// todo check here
		str = strconv.FormatFloat(fv.Float(), 'e', -1, 64)
	case reflect.String:
		str = fv.String()
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		fallthrough
	case reflect.Array:
		fallthrough
	case reflect.Chan:
		fallthrough
	case reflect.Func:
		fallthrough
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		fallthrough
	case reflect.Ptr:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Struct:
		fallthrough
	case reflect.UnsafePointer:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		log.Printf("%v not supported by now", ft)
	}
	return str
}

// BTableName only used as a place holder for tag definition for table name
type BTableName struct {
}

var tableNameType = reflect.TypeOf(BTableName{})

// BTable database sql properties defined within the sql
type BTable struct {
	SQLName    string
	Type       reflect.Type
	PrimaryKey BColumn
	Columns    []BColumn
	ColumnMap  map[string]BColumn
}

// define to store reflect in memory
var cachedDefinition = map[reflect.Type]BTable{}

// define for map definition
var cacheMutex = sync.Mutex{}

// GetDefinition get a table definition from cache or build a new one
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

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := getFields(t)

	cacheMap := map[string]int{}

	foundPrimary := false
	for i, field := range fields {
		if _, ok := cacheMap[field.Name]; ok {
			continue
		}
		if field.Type == tableNameType {
			bTable.SQLName = field.Tag.Get(sqlName)
		} else {
			column := BColumn{Field: field, FieldIndex: i}
			if field.Tag.Get(sqlIndex) != "" {
				column.Index, err = strconv.Atoi(field.Tag.Get(sqlIndex))
				if err != nil {
					return bTable, err
				}
			}
			if field.Tag.Get(sqlName) != "" {
				column.SQLName = field.Tag.Get(sqlName)
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

	bTable.ColumnMap = map[string]BColumn{}
	for _, column := range bTable.Columns {
		if _, ok := bTable.ColumnMap[column.SQLName]; ok {
			continue
		}
		bTable.ColumnMap[column.SQLName] = column
	}

	if bTable.SQLName == "" {
		bTable.SQLName = humpNamed(t.Name())
	}

	return bTable, err
}

const uppers = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func humpNamed(source string) string {
	resp := ""
	for i := 0; i < len(source); i++ {
		s := string(source[i])
		if strings.IndexAny(uppers, s) != -1 && i != 0 {
			resp += "_" + strings.ToLower(s)
		} else {
			resp += strings.ToLower(s)
		}
	}
	return resp
}

func getFields(t reflect.Type) []reflect.StructField {

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

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
