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

func (p *Postgres) GetCustomer(customerID string) (*geocloud.Customer, error) {
	c := &geocloud.Customer{}
	err := p.stmt.getCustomerByCustomerID.QueryRow(customerID).Scan(&c.ID, &c.Name)
	if err != nil {
		return c, err
	}

	return c, nil
}

func (p *Postgres) CreateCustomer(c *geocloud.Customer) error {
	_, err := p.stmt.createCustomer.Exec(c.ID, c.Name)
	if err != nil {
		return err
	}

	return nil
}
