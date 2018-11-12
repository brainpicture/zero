package zero

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kshvakov/clickhouse"
)

// ClickHouseCtx clickhause database instance
type ClickHouseCtx struct {
	DB *sql.DB
}

// ClickHouseDays days data representation
type ClickHouseDays map[int64]int64

// ClickHouse creates new database connection
func ClickHouse(dataSourceName string) *ClickHouseCtx {
	connection, err := sql.Open("clickhouse", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	if err := connection.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return nil
	}
	return &ClickHouseCtx{
		DB: connection,
	}
}

// ClickHouseParams list of values for table
type ClickHouseParams []interface{}

// ClickHouseTable is general type for table abstraction
type ClickHouseTable struct {
	name       string
	fields     []string
	prepareStr string
	rows       []ClickHouseParams
	eventChan  chan ClickHouseParams
	ctx        *ClickHouseCtx
}

// Send sends event to table
func (t *ClickHouseTable) Send(params ...interface{}) {
	t.eventChan <- params
}

func (t *ClickHouseTable) flush() {
	if len(t.rows) == 0 {
		return
	}
	tx, _ := t.ctx.DB.Begin()
	stmt, err := tx.Prepare(t.prepareStr)
	if err != nil {
		fmt.Println("CLICKHOUSE ERROR", err)
		return
	}
	for _, v := range t.rows {
		if _, err := stmt.Exec(v...); err != nil {
			fmt.Println("failed", err, v)
		} else {
			//fmt.Println("ok", v)
		}
	}
	if err := tx.Commit(); err != nil {
		fmt.Println("clickhouse FLUSH error", err)
		return
	}
	t.rows = []ClickHouseParams{}
}

// Create perfort CREATE TABLE IF NOT EXISTS query
/*func (t *ClickHouseTable) Create() {
	rows := []string{}
	for k, v := range t.fields {
		rows = append(rows, k+" "+v.(string))
	}
	sql := `CREATE TABLE IF NOT EXISTS ` + t.name + `
		(` + strings.Join(rows, ", ") + `) ENGINE = ` + t.engine

	_, err := t.ctx.DB.Exec(sql)
	if err != nil {
		fmt.Println("Clickhouse create err", err)
	}
}*/

// Watch start batching table elements
func (t *ClickHouseTable) Watch(duration time.Duration) {
	rows := []string{}
	queries := []string{}
	for _, fieldName := range t.fields {
		rows = append(rows, fieldName)
		queries = append(queries, "?")
	}
	t.prepareStr = "INSERT INTO " + t.name + " (" + strings.Join(rows, ", ") + ") VALUES (" + strings.Join(queries, ", ") + ")"
	t.eventChan = make(chan ClickHouseParams)

	go func() {
		for {
			select {
			case e := <-t.eventChan:
				t.rows = append(t.rows, e)
			case <-time.After(duration):
				t.flush()
			}
		}
	}()
}

// Count fetches an int from table
func (t *ClickHouseTable) Count(sql string) int64 {
	row := t.ctx.DB.QueryRow(sql)
	val := int64(0)
	err := row.Scan(&val)
	if err != nil {
		fmt.Println("ClickHouse count err", err)
	}
	return val
}

// DateInt allow you to fetch 2d plot data
func (t *ClickHouseTable) DateInt(sql string) ClickHouseDays {
	res := ClickHouseDays{}
	rows, err := t.ctx.DB.Query(sql)
	if err != nil {
		fmt.Println("Query error", err)
		return res
	}

	for rows.Next() {
		var date time.Time
		var val int64
		err := rows.Scan(&date, &val)
		fmt.Println("date fetched", date)
		if err != nil {
			fmt.Println("fetch error", err)
			continue
		}
		res[date.Unix()] = val
	}
	return res
}

// ListH return list of rowes from clickhouse
func (t *ClickHouseTable) ListH(sql string) []H {
	res := []H{}
	rows, err := t.ctx.DB.Query(sql)
	if err != nil {
		fmt.Println("Query error", err)
		return res
	}
	colNames, _ := rows.Columns()

	for rows.Next() {
		readCols := make([]interface{}, len(colNames))
		writeCols := make([]string, len(colNames))
		for i := range writeCols {
			readCols[i] = &writeCols[i]
		}
		err := rows.Scan(readCols...)

		if err != nil {
			fmt.Println("fetch error", err)
			continue
		}
		row := H{}
		for i, name := range colNames {
			row[name] = readCols[i]
		}
		res = append(res, row)
	}
	return res
}

// Table creates clickhouse table
func (ch *ClickHouseCtx) Table(name string, fields []string, duration time.Duration) *ClickHouseTable {
	table := &ClickHouseTable{
		name:   name,
		fields: fields,
		ctx:    ch,
	}
	table.Watch(duration)
	return table
}

// ClickHouseDate return curent date
func ClickHouseDate() clickhouse.Date {
	return clickhouse.Date(time.Now())
}

// ClickHouseCustomDate return curent date
func ClickHouseCustomDate(sec int64) clickhouse.Date {
	return clickhouse.Date(time.Unix(sec, 0))
}

// ClickHouseDateTime return curent time
func ClickHouseDateTime() clickhouse.DateTime {
	return clickhouse.DateTime(time.Now())
}

// ClickHouseCustomDateTime return curent time
func ClickHouseCustomDateTime(sec int64) clickhouse.DateTime {
	return clickhouse.DateTime(time.Unix(sec, 0))
}

// Plot plots graph
func (days ClickHouseDays) Plot(from int, to int) *Plot2D {
	res := Plot2D{
		Labels: []string{},
		Points: []int64{},
	}
	fmt.Println("DAYS", days)
	day := from
	t := time.Now()
	today := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix()
	for day <= to {
		d := today + int64(day*86400)
		val := int64(0)
		plus, ok := days[d]
		fmt.Println("date fixed", d, "got", plus, "and", ok)
		if ok {
			val += plus
		}
		res.Labels = append(res.Labels, time.Unix(d, 0).Format("Mon 2"))
		res.Points = append(res.Points, val)

		day++
	}
	return &res
}
