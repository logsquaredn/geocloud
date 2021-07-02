package das

import "database/sql"

type DasOpt func (d *Das)

func WithConnectionString(conn string) DasOpt {
	return func(d *Das) {
		d.conn = conn
	}
}

func WithDB(db *sql.DB) DasOpt {
	return func(d *Das) {
		d.db = db
	}
}
