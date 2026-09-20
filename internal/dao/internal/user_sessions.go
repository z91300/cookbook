// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// UserSessionsDao is the data access object for the table user_sessions.
type UserSessionsDao struct {
	table    string              // table is the underlying table name of the DAO.
	group    string              // group is the database configuration group name of the current DAO.
	columns  UserSessionsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler  // handlers for customized model modification.
}

// UserSessionsColumns defines and stores column names for the table user_sessions.
type UserSessionsColumns struct {
	Id               string //
	UserId           string //
	ExpiresAt        string //
	CreatedAt        string //
	UpdatedAt        string //
	AccessTokenHash  string //
	AccessExpiresAt  string //
	RefreshTokenHash string //
}

// userSessionsColumns holds the columns for the table user_sessions.
var userSessionsColumns = UserSessionsColumns{
	Id:               "id",
	UserId:           "user_id",
	ExpiresAt:        "expires_at",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
	AccessTokenHash:  "access_token_hash",
	AccessExpiresAt:  "access_expires_at",
	RefreshTokenHash: "refresh_token_hash",
}

// NewUserSessionsDao creates and returns a new DAO object for table data access.
func NewUserSessionsDao(handlers ...gdb.ModelHandler) *UserSessionsDao {
	return &UserSessionsDao{
		group:    "default",
		table:    "user_sessions",
		columns:  userSessionsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *UserSessionsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *UserSessionsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *UserSessionsDao) Columns() UserSessionsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *UserSessionsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *UserSessionsDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *UserSessionsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
