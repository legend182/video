package repo

import (
	"github.com/qinguoyi/osproxy/app/models"
	"gorm.io/gorm"
)

type ufRepo struct{}

func NewUFRepo() *ufRepo { return &ufRepo{} }

func (u *ufRepo) Create(db *gorm.DB, uf models.UserFile) (error error) {
	err := db.Create(&uf).Error
	return err
}
