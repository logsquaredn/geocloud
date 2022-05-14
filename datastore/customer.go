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
	if err := p.stmt.getCustomerByID.QueryRow(m.GetID()).Scan(&c.ID); err != nil {
		return c, err
	}

	return c, nil
}

func (p *Postgres) CreateCustomer(m geocloud.Message) (*geocloud.Customer, error) {
	if _, err := p.stmt.createCustomer.Exec(m.GetID()); err != nil {
		return nil, err
	}

	return &geocloud.Customer{
		ID: m.GetID(),
	}, nil
}
