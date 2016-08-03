package main

import (
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"

	log "github.com/Sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"

	"flag"
)

func init() {
	os.Setenv("NLS_LANG", "AMERICAN_AMERICA.AL32UTF8")

}
func main() {
	driver := flag.String("driver", "", "db driver")
	dburl := flag.String("dburl", "", "database url")
	table := flag.String("table", "", "table name")
	query := flag.String("query", "", "select query sql,the table must empty")
	limit := flag.Int64("limit", -1, "limit table rows")
	filename := flag.String("file", "export.dat", "select query sql,the table must empty")
	flag.Parse()
	var strSql string
	var count uint64
	if len(*query) > 0 {
		strSql = *query
	} else {
		if len(*table) == 0 {
			log.Panic("table is empty")
		}
		if *limit >= 0 {
			switch *driver {
			case "oci8":
				strSql = fmt.Sprintf("select * from %s where rownum<=%d", *table, *limit)
			case "postgres":
				strSql = fmt.Sprintf("select * from %s limit %d", *table, *limit)
			case "mysql":
				strSql = fmt.Sprintf("select * from %s limit %d", *table, *limit)
			default:
				log.Panic("invalid driver", *driver)
			}
		} else {
			strSql = fmt.Sprintf("select * from %s", *table)
		}
	}
	db, err := sqlx.Open(*driver, *dburl)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	row := db.QueryRow(fmt.Sprintf("select count(*) from (%s) cnt_sql", strSql))
	if err = row.Scan(&count); err != nil {
		log.Panic(err)
	}
	file, err := os.Create(*filename)
	if err != nil {
		log.WithFields(log.Fields{
			"file": filename,
		}).Panic(err)
	}
	defer file.Close()

	rows, err := db.Queryx(strSql)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()
	var rn uint64 = 0
	cols, err := rows.Columns()
	if err != nil {
		log.Panic(err)
	}
	enc := gob.NewEncoder(file)
	if err = enc.Encode(cols); err != nil {
		log.Panic(err)
	}
	startTime := time.Now()
	preTime := startTime
	var preCount uint64
	for rows.Next() {
		rn++
		vars, err := rows.SliceScan()
		if err != nil {
			log.WithFields(log.Fields{
				"no": rn,
			}).Panic(err)
		}
		if err = enc.Encode(vars); err != nil {
			log.WithFields(log.Fields{
				"no": rn,
			}).Panic(err)
		}
		if time.Since(preTime).Seconds() > 15 {
			log.WithFields(log.Fields{
				"no":  rn,
				"op":  fmt.Sprintf("%.0f", float64(rn-preCount)/time.Since(preTime).Seconds()),
				"prg": fmt.Sprintf("%.2f", float64(rn)/float64(count)*100),
			}).Info("progress")
			preTime = time.Now()
			preCount = rn
		}
	}
	log.WithFields(log.Fields{
		"sql":  strSql,
		"file": *filename,
		"rows": rn,
	}).Info("finish")
}
