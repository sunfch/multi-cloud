package tidbclient

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opensds/multi-cloud/s3/pkg/datastore/yig/helper"
	"os"
	"time"
)

const MAX_OPEN_CONNS = 8196

type TidbClient struct {
	Client *sql.DB
}

func NewTidbClient() *TidbClient {
	cli := &TidbClient{}
	conn, err := sql.Open("mysql", helper.CONFIG.TidbInfo)
	if err != nil {
		os.Exit(1)
	}
	conn.SetMaxIdleConns(256)
	conn.SetMaxOpenConns(MAX_OPEN_CONNS)
	conn.SetConnMaxLifetime(300 * time.Second)
	cli.Client = conn
	return cli
}
