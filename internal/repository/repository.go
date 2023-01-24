package repository

import (
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/table"
)

type Repository[T any] interface {
	Insert(model *T) error
	GetOne(model *T) error
	GetAll(model *T) (*[]T, error)
	Delete(model *T) error
	Session() *gocqlx.Session
	TableDefinition() *table.Table
	Query(queryKey string, getter statementGetter) *gocqlx.Queryx
}

type statementGetter = func() (stmt string, names []string)
type queries = map[string]*gocqlx.Queryx

type CassandraRepository[T any] struct {
	session *gocqlx.Session
	table   *table.Table
	queries queries
}

func New[T any](session *gocqlx.Session, table *table.Table) CassandraRepository[T] {
	queries := queries{}

	return CassandraRepository[T]{
		session,
		table,
		queries,
	}
}

func (r CassandraRepository[T]) lazyGetQuery(queryKey string, getter statementGetter) *gocqlx.Queryx {
	if r.queries[queryKey] != nil {
		return r.queries[queryKey]
	}
	r.queries[queryKey] = r.session.Query(getter())
	return r.queries[queryKey]
}

func (r CassandraRepository[T]) Insert(model *T) error {
	q := r.lazyGetQuery("Insert", r.table.Insert).BindStruct(*model)

	return q.Exec()
}

func (r CassandraRepository[T]) GetOne(model *T) error {
	q := r.lazyGetQuery("Get", toStatementGetter(r.table.Get)).BindStruct(*model)

	return q.Get(model)
}

func (r CassandraRepository[T]) GetAll(model *T) (*[]T, error) {
	result := new([]T)
	q := r.lazyGetQuery("Select", toStatementGetter(r.table.Select)).BindStruct(*model)

	err := q.Select(result)

	return result, err
}

func (r CassandraRepository[T]) Delete(model *T) error {
	q := r.lazyGetQuery("Delete", toStatementGetter(r.table.Delete)).BindStruct(*model)

	return q.Exec()
}

func (r CassandraRepository[T]) Session() *gocqlx.Session {
	return r.session
}

func (r CassandraRepository[T]) TableDefinition() *table.Table {
	return r.table
}

func (r CassandraRepository[T]) Query(queryKey string, getter statementGetter) *gocqlx.Queryx {
	return r.lazyGetQuery(queryKey, getter)
}

func toStatementGetter(fn func(...string) (string, []string)) statementGetter {
	return func() (string, []string) {
		return fn()
	}
}
