package sqlhelper

import (
	"context"
	"database/sql"
	"errors"
	"github.com/cocotyty/sqlhelper/internel"
)

// use for validate interface
var _ SQLHelper = (*sqlHelper)(nil)

func New(db *sql.DB) SQLHelper {
	return &sqlHelper{
		db:           db,
		Scanner:      internel.GlobalScanner,
		SQLGenerator: internel.GlobalSQLGenerator,
	}
}

func NewSQLHelper(db operator, Scanner *internel.RowsScanner, SQLGenerator *internel.SQLGenerator) *sqlHelper {
	return &sqlHelper{
		db:           db,
		Scanner:      Scanner,
		SQLGenerator: SQLGenerator,
	}
}

type sqlHelper struct {
	db           operator
	Scanner      *internel.RowsScanner
	SQLGenerator *internel.SQLGenerator
}

func (s *sqlHelper) execute(ctx context.Context, db operator, sqlstr string, args ...interface{}) (int64, int64, error) {
	result, err := db.ExecContext(ctx, sqlstr, args...)
	if err != nil {
		return 0, 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	num, err := result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return id, num, nil
}

// 插入数据
func (s *sqlHelper) InsertContext(ctx context.Context, sqlstr string, args ...interface{}) (int64, error) {

	id, _, err := s.execute(ctx, s.db, sqlstr, args...)

	return id, err
}

// 插入数据
func (s *sqlHelper) InsertObject(ctx context.Context, object interface{}) (int64, error) {
	sqlStr, args, err := s.SQLGenerator.PrepareInsert(object)
	if err != nil {
		return 0, err
	}
	return s.InsertContext(ctx, sqlStr, args...)
}

// 更新对象所有属性 但不更新对象的ID
func (s *sqlHelper) UpdateObjectWhere(ctx context.Context, object interface{}, where string, optionArgs ...interface{}) (int64, error) {
	if where == "" {
		return 0, errors.New("forbidden operation: empty `where` param")
	}
	sqlStr, args, err := s.SQLGenerator.PrepareUpdate(object)
	if err != nil {
		return 0, err
	}
	sqlStr += " WHERE " + where
	args = append(args, optionArgs...)
	return s.UpdateContext(ctx, sqlStr, args...)
}

func (s *sqlHelper) UpdateObjectByID(ctx context.Context, object interface{}) (int64, error) {
	sqlStr, args, err := s.SQLGenerator.PrepareUpdateByID(object)
	if err != nil {
		return 0, err
	}
	return s.UpdateContext(ctx, sqlStr, args...)
}

// 删除数据
func (s *sqlHelper) DeleteContext(ctx context.Context, sqlstr string, args ...interface{}) (int64, error) {
	_, num, err := s.execute(ctx, s.db, sqlstr, args...)

	return num, err
}

// QueryContext 查询指定sqlstr的语句 并使用args来填充statement，将返回结果反序列化到指针ptr中
// ptr 支持的类型有:
// 1. 指向结构体的指针 如 var u *User
// 2. 指向原始类型的指针 如 var x *int 原始类型包括 bool int float string []byte time.Time 和所有sql.Scanner接口的实现类型
// 3. 指向结构体的slice的指针 如 var list *[]User
// 4. 指向结构体指针的slice的指针 如 var list *[]*User
// 5. 指向原始类型的slice的指针 如 var list *[]int
// 警告：若ptr类型为单行值如结构体指针等时，若未查询到结果，返回值err为sql.ErrNoRows
// 此时需要判断error是否为sql.ErrNoRows这种错误，若为这种错误，则说明未查询到，业务方自行处理
// 若ptr类型为执行slice的指针，则不会出现error为sql.ErrNoRows的情况
// context参数 可适用于opentracing 不可为nil
func (s *sqlHelper) QueryContext(ctx context.Context, ptr interface{}, sqlstr string, args ...interface{}) error {

	rows, err := s.db.QueryContext(ctx, sqlstr, args...)
	if err != nil {
		return err
	}
	return s.Scanner.Scan(rows, ptr)
}

//
func (s *sqlHelper) SelectFrom(ctx context.Context, ptr interface{}, subSQL string, args ...interface{}) error {
	selectSQL, err := s.SQLGenerator.PrepareSelectFrom(ptr)
	if err != nil {
		return err
	}
	return s.QueryContext(ctx, ptr, selectSQL+subSQL, args...)
}

// 更新数据
func (s *sqlHelper) UpdateContext(ctx context.Context, sqlstr string, args ...interface{}) (int64, error) {
	_, num, err := s.execute(ctx, s.db, sqlstr, args...)

	return num, err
}
