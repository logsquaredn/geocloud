package datastore

import (
	_ "embed"

	"github.com/logsquaredn/geocloud"
)

var (
	//go:embed psql/queries/get_customer_by_id.sql
	getCustomerByIDSQL string

	//go:embed psql/execs/create_customer.sql
	createCustomerSQL string
)

func (p *Postgres) GetCustomer(m geocloud.Message) (*geocloud.Customer, error) {
	c := &geocloud.Customer{}
	err := p.stmt.getCustomerByID.QueryRow(m.GetID()).Scan(&c.ID)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (p *Postgres) CreateCustomer(m geocloud.Message) (*geocloud.Customer, error) {
	c := &geocloud.Customer{}
	err := p.stmt.createCustomer.QueryRow(m.GetID()).Scan(&c.ID)
	if err != nil {
		return nil, err
	}

	return c, nil
}
