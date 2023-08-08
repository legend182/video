package repo

import (
	"errors"
	"github.com/qinguoyi/osproxy/app/models"
	"gorm.io/gorm"
)

type userRepo struct{}

func NewUserRepo() *userRepo { return &userRepo{} }

func (u *userRepo) CheckUserExist(db *gorm.DB, username string) (error error) {
	var user models.User
	tx := db.Where("username = ?", username).Find(&user)
	if tx.RowsAffected != 0 {
		return errors.New("用户已存在")
	}
	return nil
}
func (u *userRepo) Create(db *gorm.DB, user *models.User) (error error) {
	err := db.Create(user).Error
	return err
}
func (u *userRepo) SelectByName(db *gorm.DB, name string) (user models.User, err error) {
	err = db.Where("username=?", name).First(&user).Error
	return
}
