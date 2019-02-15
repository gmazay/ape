package db

import (
	"ape/conf"
	//"fmt"
	//"os"
	"database/sql"
	//"github.com/kpango/glg"
	"ape/glg"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func Connect(cfg *conf.Config, dsn string) *sql.DB {
	var db_uri string
	switch cfg.Dsn[dsn].Type {
	case "mysql":
		db_uri = cfg.Dsn[dsn].User + ":" + cfg.Dsn[dsn].Pass + "@tcp(" + cfg.Dsn[dsn].Host + ")/" + cfg.Dsn[dsn].Dbname + cfg.Dsn[dsn].Params
	case "postgres":
		db_uri = cfg.Dsn[dsn].Type + "://" + cfg.Dsn[dsn].User + ":" + cfg.Dsn[dsn].Pass + "@" + cfg.Dsn[dsn].Host + "/" + cfg.Dsn[dsn].Dbname + cfg.Dsn[dsn].Params
	}
	//fmt.Println(db_uri)
	db, err := sql.Open(cfg.Dsn[dsn].Type, db_uri)
	if err != nil {
		glg.Errorf("DB connect error: ", err)
		panic(err)
	}
	if err = db.Ping(); err != nil {
		glg.Errorf("DB ping error: ", err)
		panic(err)
	}
	//defer db.Close()
	return db
}

func Close(dbh *sql.DB) {
	dbh.Close()
}

func FetchAll(dbh *sql.DB, query string, args ...interface{}) ([][]string, error) {
	res := make([][]string, 0)
	rows, err := dbh.Query(query, args...)
	if err != nil {
		//glg.Errorf("db: Query error: %s", err)
		return res, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		//glg.Errorf("db: Failed to get columns: %s", err)
		return res, err
	}

	destPtrs := make([]interface{}, len(cols))

	for rows.Next() {
		dest := make([]string, len(cols))

		for i, _ := range dest {
			destPtrs[i] = &dest[i]
		}

		err := rows.Scan(destPtrs...)
		if err != nil {
			//glg.Errorf("db: rows.Scan error: %s", err)
			//continue
			return res, err
		}
		res = append(res, dest)
	}

	return res, err
}

func FetchAll2(dbh *sql.DB, query string, args []interface{}) ([][]string, error) {
	res := make([][]string, 0)
	rows, err := dbh.Query(query, args...)
	if err != nil {
		//glg.Errorf("db: Query error: %s", err)
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		//glg.Errorf("db: Failed to get columns: %s", err)
		return nil, err
	}

	destPtrs := make([]interface{}, len(cols))

	for rows.Next() {
		dest := make([]string, len(cols))

		for i, _ := range dest {
			destPtrs[i] = &dest[i]
		}

		err := rows.Scan(destPtrs...)
		if err != nil {
			//glg.Errorf("db: rows.Scan error: %s", err)
			continue
		}
		res = append(res, dest)
	}

	return res, err
}

// Simple, but need number of columns(res_len) in param
func FetchRowFl(dbh *sql.DB, res_len int, query string, args ...interface{}) ([]string, error) {
	resPtrs := make([]interface{}, res_len)
	res := make([]string, res_len)
	for i, _ := range res {
		resPtrs[i] = &res[i]
	}

	err := dbh.QueryRow(query, args...).Scan(resPtrs...)
	if err != nil {
		return nil, err
	}

	return res, err
}

func FetchRow(dbh *sql.DB, query string, args ...interface{}) ([]string, error) {
	rows, err := dbh.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	resPtrs := make([]interface{}, len(cols))

	if rows.Next() {
		res := make([]string, len(cols))

		for i, _ := range res {
			resPtrs[i] = &res[i]
		}

		err = rows.Scan(resPtrs...)
		if err != nil {
			glg.Errorf("db: rows.Scan error: %s", err)
		}
		return res, err
	}

	return nil, err
}

func Do(dbh *sql.DB, query string, args ...interface{}) error {
	_, err := dbh.Query(query, args...)
	return err
}
