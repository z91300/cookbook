// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// AiGenerationsDao is the data access object for the table ai_generations.
type AiGenerationsDao struct {
	table    string               // table is the underlying table name of the DAO.
	group    string               // group is the database configuration group name of the current DAO.
	columns  AiGenerationsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler   // handlers for customized model modification.
}

// AiGenerationsColumns defines and stores column names for the table ai_generations.
type AiGenerationsColumns struct {
	Id        string //
	UserId    string //
	RecipeId  string //
	Model     string //
	Prompt    string //
	RawOutput string //
	Status    string //
	Error     string //
	CreatedAt string //
}

// aiGenerationsColumns holds the columns for the table ai_generations.
var aiGenerationsColumns = AiGenerationsColumns{
	Id:        "id",
	UserId:    "user_id",
	RecipeId:  "recipe_id",
	Model:     "model",
	Prompt:    "prompt",
	RawOutput: "raw_output",
	Status:    "status",
	Error:     "error",
	CreatedAt: "created_at",
}

// NewAiGenerationsDao creates and returns a new DAO object for table data access.
func NewAiGenerationsDao(handlers ...gdb.ModelHandler) *AiGenerationsDao {
	return &AiGenerationsDao{
		group:    "default",
		table:    "ai_generations",
		columns:  aiGenerationsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *AiGenerationsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *AiGenerationsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *AiGenerationsDao) Columns() AiGenerationsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *AiGenerationsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *AiGenerationsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *AiGenerationsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
