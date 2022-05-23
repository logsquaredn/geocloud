package geocloud

type Customer struct {
	ID string `json:"id,omitempty"`
}

var _ Message = (*Customer)(nil)

func (c *Customer) GetID() string {
	return c.ID
}
