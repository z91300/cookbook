// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// RecipesDao is the data access object for the table recipes.
type RecipesDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  RecipesColumns     // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// RecipesColumns defines and stores column names for the table recipes.
type RecipesColumns struct {
	Id                string //
	UserId            string //
	Title             string //
	Summary           string //
	CoverAttachmentId string //
	Tips              string //
	Ingredients       string //
	Tools             string //
	Steps             string //
	MealMask          string //
	Calories          string //
	Difficulty        string //
	Servings          string //
	CookMinutes       string //
	Source            string //
	AiModel           string //
	ReviewStatus      string //
	IsDeleted         string //
	CreatedAt         string //
	UpdatedAt         string //
}

// recipesColumns holds the columns for the table recipes.
var recipesColumns = RecipesColumns{
	Id:                "id",
	UserId:            "user_id",
	Title:             "title",
	Summary:           "summary",
	CoverAttachmentId: "cover_attachment_id",
	Tips:              "tips",
	Ingredients:       "ingredients",
	Tools:             "tools",
	Steps:             "steps",
	MealMask:          "meal_mask",
	Calories:          "calories",
	Difficulty:        "difficulty",
	Servings:          "servings",
	CookMinutes:       "cook_minutes",
	Source:            "source",
	AiModel:           "ai_model",
	ReviewStatus:      "review_status",
	IsDeleted:         "is_deleted",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewRecipesDao creates and returns a new DAO object for table data access.
func NewRecipesDao(handlers ...gdb.ModelHandler) *RecipesDao {
	return &RecipesDao{
		group:    "default",
		table:    "recipes",
		columns:  recipesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *RecipesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *RecipesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *RecipesDao) Columns() RecipesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *RecipesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *RecipesDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *RecipesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
