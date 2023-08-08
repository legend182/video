package models

type UserFile struct {
	Ud     int   `gorm:"column:id;primaryKey;not null;autoIncrement;comment:自增ID" json:"id,string" db:"id"`
	UserId int64 `gorm:"column:userid" json:"userid" db:"userid"`
	FileId int64 `gorm:"column:fileid" json:"fileid" db:"fileid"`
}
