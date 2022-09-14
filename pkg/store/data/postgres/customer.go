package postgres

import (
	_ "embed"
	"strings"

	"github.com/google/uuid"
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
	if err := d.stmt.getCustomerByID.QueryRow(id).Scan(&c.Id, &c.ApiKey, &c.Email); err != nil {
		return c, err
	}

	return c, nil
}

func (d *Datastore) CreateCustomer(id string, email string) (*api.Customer, error) {
	apikey := strings.ReplaceAll(uuid.New().String(), "-", "")
	if _, err := d.stmt.createCustomer.Exec(id, apikey, email); err != nil {
		return nil, err
	}

	return &api.Customer{
		Id:     id,
		ApiKey: apikey,
		Email:  email,
	}, nil
}
