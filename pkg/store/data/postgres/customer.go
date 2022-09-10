package postgres

import (
	_ "embed"

	"github.com/logsquaredn/rototiller/pkg/api"
)

var (
	//go:embed sql/queries/get_customer_by_id.sql
	getCustomerByIDSQL string

	//go:embed sql/execs/create_customer.sql
	createCustomerSQL string
)

func (d *Datastore) GetCustomer(id string) (*api.Customer, error) {
	c := &api.Customer{}
	if err := d.stmt.getCustomerByID.QueryRow(id).Scan(&c.Id); err != nil {
		return c, err
	}

	return c, nil
}

func (d *Datastore) CreateCustomer(id string) (*api.Customer, error) {
	if _, err := d.stmt.createCustomer.Exec(id); err != nil {
		return nil, err
	}

	return &api.Customer{
		Id: id,
	}, nil
}
