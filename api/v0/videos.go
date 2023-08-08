package v0

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/qinguoyi/osproxy/app/models"
	"github.com/qinguoyi/osproxy/app/pkg/utils"
	"github.com/qinguoyi/osproxy/app/pkg/web"
	"github.com/qinguoyi/osproxy/bootstrap/plugins"
	"strconv"
)

// Videos 获取列表
//
//	@Summary		获取视频uid
//	@Description	获取视频
//	@Tags			列表
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	web.Response
//	@Router			/api/storage/v0/videos [get]
func Videos(c *gin.Context) {

	var right int64
	lgRedis := new(plugins.LangGoRedis).NewRedis()
	get := lgRedis.HGet(c, utils.VideoHash, "root").Val()
	var index int64
	if get == "" {
		index = 0
	} else {
		var err error
		index, err = strconv.ParseInt(get, 10, 64)
		if err != nil {
			panic(err)
		}
	}
	lLen := lgRedis.LLen(c, utils.Video).Val()
	if index+utils.VideoCount < lLen-1 {
		right = index + utils.VideoCount
		lgRedis.HSet(c, utils.VideoHash, "root", right+1)
	} else {
		right = lLen - 1
		lgRedis.HSet(c, utils.VideoHash, "root", 0)
	}
	lRange := lgRedis.LRange(c, utils.Video, index, right)
	if lRange.Err() != nil {
		lgLogger.WithContext(c).Error("获取视频出错")
	}
	result, err := lRange.Result()
	if err != nil {
		panic(err)
	}
	var data []models.VideoInfo
	for _, jsonData := range result {
		var video models.VideoInfo
		err := json.Unmarshal([]byte(jsonData), &video)
		if err != nil {
			panic(err)
		}
		data = append(data, video)
	}
	web.Success(c, data)
}
