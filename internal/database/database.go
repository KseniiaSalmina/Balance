package database

import (
	"errors"
)

//OrderBy can be date or amount
type OrderBy string

const (
	OrderByDate   OrderBy = "date"
	OrderByAmount OrderBy = "amount"
)

//Order can be DESC or ASC
type Order string

const (
	Desc Order = "DESC"
	Asc  Order = "ASC"
)

var UserDoesNotExistErr error = errors.New("user does not exist")
