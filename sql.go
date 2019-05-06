package sqlhelper

import (
	"context"
	"database/sql"
)

// SQLHelper 执行SQL 方便转换为对象操作
type SQLHelper interface {
	InsertContext(ctx context.Context, sqlstr string, args ...interface{}) (int64, error)
	InsertObject(ctx context.Context, object interface{}) (int64, error)
	DeleteContext(ctx context.Context, sqlstr string, args ...interface{}) (int64, error)
	SelectFrom(ctx context.Context, ptr interface{}, subSQL string, args ...interface{}) error
	UpdateContext(ctx context.Context, sqlstr string, args ...interface{}) (int64, error)
	UpdateObjectWhere(ctx context.Context, object interface{}, where string, optionArgs ...interface{}) (int64, error)
	UpdateObjectByID(ctx context.Context, object interface{}) (int64, error)
	QueryContext(ctx context.Context, ptr interface{}, sqlstr string, args ...interface{}) error
}

// transaction 可以提交回滚的事务
type transaction interface {
	Commit() error
	Rollback() error
}

// sqlTransaction 数据库事务执行的具柄
type sqlTransaction struct {
	SQLHelper
	transaction
}

func newTransaction(helper SQLHelper, rc transaction) *sqlTransaction {
	return &sqlTransaction{
		SQLHelper:   helper,
		transaction: rc,
	}
}

// operator 处理数据库实际的执行
type operator interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Close() error
}
