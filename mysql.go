package zero

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql" // using mysql in library
)

// MySQL is main class for managing db
type MySQL struct {
	db *sql.DB
}

func (c *MySQL) scanFields(prefix string, obj interface{}, values []interface{}, queryTypes, queryQueries []string, len int) ([]interface{}, []string, []string, int) {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	numFields := v.NumField()
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		tag := field.Tag.Get("mysql")
		if tag != "" {
			vfield := v.Field(i)
			if vfield.Kind() == reflect.Ptr {
				vfield = vfield.Elem()
			}
			if vfield.Kind() == reflect.Struct {
				values, queryTypes, queryQueries, len = c.scanFields(prefix+tag+"_", vfield.Interface(), values, queryTypes, queryQueries, len)
			} else {
				queryTypes = append(queryTypes, prefix+tag)
				queryQueries = append(queryQueries, "?")
				values[len] = vfield.Interface()
				len++
			}
		}
	}
	return values, queryTypes, queryQueries, len
}

func (c *MySQL) inspect(obj interface{}, addValues []interface{}) ([]interface{}, []string, []string) {
	var values []interface{}
	queryTypes := []string{}
	queryQueries := []string{}
	values = make([]interface{}, 100) // 100 fields is maximum
	len := 0
	if reflect.ValueOf(obj).Kind() == reflect.Map {
		objMap := obj.(H)
		for k, v := range objMap {
			queryTypes = append(queryTypes, k)
			queryQueries = append(queryQueries, "?")
			values[len] = v
			len++
		}
	} else {
		values, queryTypes, queryQueries, len = c.scanFields("", obj, values, queryTypes, queryQueries, len)
	}
	if addValues != nil {
		for _, v := range addValues {
			values[len] = v
			len++
		}
	}
	values = values[0:len]
	return values, queryTypes, queryQueries
}

// Insert inserts an object to the database
func (c *MySQL) Insert(table string, obj interface{}) error {
	values, queryTypes, queryQueries := c.inspect(obj, nil)

	sql := "INSERT INTO " + table + " (" + strings.Join(queryTypes, ", ") + ") VALUES (" + strings.Join(queryQueries, ", ") + ")"
	_, err := c.db.Exec(sql, values...)
	if err != nil {
		fmt.Println("error:", sql, err)
	} else {
		fmt.Println("ok:", sql)
	}
	return err
}

// Replace replaces an object in the database. Be aware that replace perform delete and insert if unique index
func (c *MySQL) Replace(table string, obj interface{}) error {
	values, queryTypes, queryQueries := c.inspect(obj, nil)

	sql := "REPLACE INTO " + table + " (" + strings.Join(queryTypes, ", ") + ") VALUES (" + strings.Join(queryQueries, ", ") + ")"
	_, err := c.db.Exec(sql, values...)
	if err != nil {
		fmt.Println("error:", sql, err)
	} else {
		fmt.Println("ok:", sql)
	}
	return err
}

// Update updates any Row in db
func (c *MySQL) Update(table string, obj interface{}, checkField string, checkValues ...interface{}) {
	values, queryTypes, _ := c.inspect(obj, checkValues)
	for k, v := range queryTypes {
		queryTypes[k] = v + " = ?"
	}
	sql := "UPDATE " + table + " SET " + strings.Join(queryTypes, ", ") + " WHERE " + checkField
	fmt.Println("sql", sql)
	_, err := c.db.Exec(sql, values...)
	fmt.Println("error", err)
}

// QueryRow allow fetch any data
func (c *MySQL) QueryRow(sql string, values ...interface{}) *sql.Row {
	return c.db.QueryRow(sql, values...)
}

// Connect connects to db
func (c *MySQL) Connect(initStr string) {
	db, err := sql.Open("mysql", initStr)
	if err != nil {
		fmt.Println("MYSQL connect error", err)
	}
	c.db = db
}
