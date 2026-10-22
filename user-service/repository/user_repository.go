package repository

import (
	"example.com/go-commerce-platform/user-service/models"
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

type UserReppository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserReppository {
	return &UserReppository{
		DB: db,
	}
}

func (r *UserReppository) CreateUser(ctx context.Context, user *models.User) error {
	err := user.Insert(ctx, r.DB, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}
func (r *UserReppository) GetUserById(ctx context.Context, id int) (*models.User, error) {
	user, err := models.FindUser(ctx, r.DB, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (r *UserReppository) DeleteUserBYId(ctx context.Context, id int) error {
	user, err := models.FindUser(ctx, r.DB, id)
	if err != nil {
		return err
	}

	_, err = user.Delete(ctx, r.DB)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserReppository) GetAllUsers(ctx context.Context, name string, limit int, offset int) ([]*models.User, error) {
	queryMods := []qm.QueryMod{
		qm.OrderBy("id ASC"),
	}
	if name != "" {
		queryMods = append(queryMods, qm.Where("name ILIKE ?", "%"+name+"%"))
	}
	queryMods = append(queryMods, qm.Limit(limit), qm.Offset(offset))

	users, err := models.Users(queryMods...).All(ctx, r.DB)
	if err != nil {
		return nil, err
	}
	return users, nil

}
func (r *UserReppository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := models.Users(qm.Where("email = ?", email)).One(ctx, r.DB)
	if err != nil {
		return nil, err
	}
	return user, nil
}
