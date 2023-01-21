package api

type listQuery struct {
	Offset int `form:"offset,default=0"`
	Limit  int `form:"limit,default=10"`
}
