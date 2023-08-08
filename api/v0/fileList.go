package v0

import (
	"github.com/gin-gonic/gin"
	"github.com/qinguoyi/osproxy/app/models"
	"github.com/qinguoyi/osproxy/app/pkg/web"
	"github.com/qinguoyi/osproxy/bootstrap/plugins"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// GetFileList DownloadHandler   获取列表
//
//	@Summary		获取列表
//	@Description	获取列表
//	@Tags			列表
//	@Accept			application/json
//	@Param			page		query	string	true	"page"
//	@Param			pageSize	query	string	true	"pageSize"
//	@Produce		application/json
//	@Success		200	{object}	web.Response
//	@Router			/api/storage/v0/fileList [get]
func GetFileList(c *gin.Context) {
	var files []models.MetaDataInfo
	lgDB := new(plugins.LangGoDB).Use("default").NewDB()
	lgDB.Scopes(FileStatus(), PageNate(c.Request)).Find(&files)
	var file1 []models.DataInfo
	for _, f := range files {
		file := models.DataInfo{
			UID:       f.UID,
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		}
		file1 = append(file1, file)
	}
	web.Success(c, file1)
}
func FileStatus() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status =?", 1)
	}
}
func PageNate(r *http.Request) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page <= 0 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(q.Get("pageSize"))
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
