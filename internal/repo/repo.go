package repo

import (
	"github.com/jmoiron/sqlx"
	"github.com/xligenda/ods-servers/internal/structs"
)

type IDsConstraint interface {
	string | int
}

type StructsConstraint[I IDsConstraint] interface {
	structs.Server
	GetID() I
}

type GenericRepository[I IDsConstraint, T StructsConstraint[I]] struct {
	db        *sqlx.DB
	tableName string
}

type QueryOptions struct {
	OrderBy string
	Limit   int
	Offset  int
}

type Filter struct {
	Field string
	// =, !=, >, <, >=, <=, LIKE, IN, NOT IN, OR
	Operator string
	Value    any
}

func NewRepository[I IDsConstraint, T StructsConstraint[I]](db *sqlx.DB, tableName string) *GenericRepository[I, T] {
	return &GenericRepository[I, T]{
		db:        db,
		tableName: tableName,
	}
}
