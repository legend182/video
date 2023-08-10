package models

import "time"

type CommentInfo struct {
	ID            int        `gorm:"column:id;primaryKey;not null;autoIncrement;comment:自增ID"`
	UID           int64      `gorm:"column:uid;primaryKey;not null;comment:唯一ID"`
	Content       string     `gorm:"column:content;not null"`
	IsDelete      bool       `gorm:"column:is_delete"`
	UserId        int64      `gorm:"column:user_id;not null"`
	VideoId       int64      `gorm:"column:video_id;not null"`
	LikeCount     int        `gorm:"column:like_count"`
	RootCommentId int64      `gorm:"column:root_commentId"`
	ToUserId      int64      `gorm:"column:to_user_id"`
	CreateTime    *time.Time `gorm:"column:create_time;not null"`
}
type SendReq struct {
	Content string `json:"content"`
	VideoId int64  `json:"videoId,string"`
	//UserId        int64  `json:"userId,string"`
	RootCommentId int64 `json:"rootCommentId,string"`
	ToUserId      int64 `json:"toUserId,string"`
}
type ClickReq struct {
	CommentId int64 `json:"commentId,string"`
	UserId    int64 `json:"userId,string"`
	// 0点赞 1取消点赞
	IsLove int `json:"isLove"`
}
type ShowComment struct {
	Content       string        `json:"content"`
	UserId        int64         `json:"userId,string"`
	LikeCount     int           `json:"likeCount"`
	RootCommentId int64         `json:"rootCommentId,string"`
	ToUserId      int64         `json:"toUserId,string"`
	CreateTime    *time.Time    `json:"createTime"`
	SonComment    []ShowComment `json:"showComment"`
}
