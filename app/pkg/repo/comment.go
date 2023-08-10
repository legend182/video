package repo

import (
	"github.com/qinguoyi/osproxy/app/models"
	"gorm.io/gorm"
)

type comRepo struct{}

func NewComRepo() *comRepo { return &comRepo{} }
func (c *comRepo) Create(db *gorm.DB, uf models.CommentInfo) (error error) {
	error = db.Create(&uf).Error
	return error
}

// SelectByUid 获取根评论
func (c *comRepo) SelectByUid(db *gorm.DB, uid int64) (req []models.CommentInfo, error error) {
	tx := db.Where("video_id=? And root_commentId =?", uid, 0).Find(&req)
	return req, tx.Error
}

// SelectByParent 获取子评论
func (c *comRepo) SelectByParent(db *gorm.DB, uid int64) (req []models.CommentInfo, error error) {
	tx := db.Where("root_commentId =?", uid).Find(&req)
	return req, tx.Error
}
