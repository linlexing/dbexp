package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"

	log "github.com/Sirupsen/logrus"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"

	"flag"

	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var cfg *Config

func init() {
	os.Setenv("NLS_LANG", "AMERICAN_AMERICA.AL32UTF8")

}
func main() {
	cfg = new(Config)
	bs, err := ioutil.ReadFile("config.yaml")
	if err == nil {
		if err = yaml.Unmarshal(bs, cfg); err != nil {
			log.Panic(err)
		}
	} else {
		driver := flag.String("driver", "", "db driver")
		dburl := flag.String("dburl", "", "database url")
		table := flag.String("table", "", "table name")
		query := flag.String("query", "", "select query sql,the table must empty")
		limit := flag.Int64("limit", -1, "limit table rows")
		filename := flag.String("file", "export.dat", "output file name")
		outfmt := flag.String("fmt", "gob", "output file format. gob/flat")
		flag.Parse()
		cfg.Driver = *driver
		cfg.DBUrl = *dburl
		cfg.Table = *table
		cfg.Query = *query
		cfg.Limit = *limit
		cfg.Filename = *filename
		cfg.Outfmt = *outfmt
	}

	var strSql string
	var count uint64
	if len(cfg.Query) > 0 {
		strSql = cfg.Query
	} else {
		if len(cfg.Table) == 0 {
			log.Panic("table is empty")
		}
		if cfg.Limit >= 0 {
			switch cfg.Driver {
			case "oci8":
				strSql = fmt.Sprintf("select * from %s where rownum<=%d", cfg.Table, cfg.Limit)
			case "postgres":
				strSql = fmt.Sprintf("select * from %s limit %d", cfg.Table, cfg.Limit)
			case "mysql":
				strSql = fmt.Sprintf("select * from %s limit %d", cfg.Table, cfg.Limit)
			default:
				log.Panic("invalid driver", cfg.Driver)
			}
		} else {
			strSql = fmt.Sprintf("select * from %s", cfg.Table)
		}
	}
	var out outEncode
	switch cfg.Outfmt {
	case "gob":
		out = new(outGob)
	case "flat":
		out = new(outFlat)
	default:
		log.Panic("invalid outfmt,", cfg.Outfmt)
	}
	db, err := sqlx.Open(cfg.Driver, cfg.DBUrl)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	row := db.QueryRow(fmt.Sprintf("select count(*) from (%s) cnt_sql", strSql))
	if err = row.Scan(&count); err != nil {
		log.Panic(err)
	}
	if err := out.Open(cfg.Filename); err != nil {
		log.WithFields(log.Fields{
			"file": cfg.Filename,
		}).Panic(err)
	}
	defer out.Close()

	rows, err := db.Queryx(strSql)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Panic(err)
	}
	if len(cfg.FieldSize) > 0 && len(cfg.FieldSize) != len(cols) {
		log.Panic(fmt.Sprintf("error col num %d not equ fieldsize %d", len(cols), len(cfg.FieldSize)))
	}
	var rn uint64
	if err = out.WriteTitle(cols); err != nil {
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
		if err = out.WriteLine(vars); err != nil {
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
		"file": cfg.Filename,
		"rows": rn,
	}).Info("finish")
}
