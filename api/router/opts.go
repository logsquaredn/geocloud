package router

import (
	"database/sql"

	"github.com/logsquaredn/geocloud/shared/das"
)

type RouterOpt func (r *Router)

func WithConnectionString(conn string) RouterOpt {
	return func(r *Router) {
		r.conn = conn
	}
}

func WithDas(das *das.Das) RouterOpt {
	return func(r *Router) {
		r.das = das
	}
}

func WithDB(db *sql.DB) RouterOpt {
	return func(r *Router) {
		r.db = db
	}
}
