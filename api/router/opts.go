package router

import (
	"database/sql"

	"github.com/logsquaredn/geocloud/shared/das"
	"github.com/logsquaredn/geocloud/shared/oas"
)

type RouterOpt func(r *Router)

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

func WithOas(oas *oas.Oas) RouterOpt {
	return func(r *Router) {
		r.oas = oas
	}
}
