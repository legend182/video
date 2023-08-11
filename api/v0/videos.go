package v0

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/qinguoyi/osproxy/app/models"
	"github.com/qinguoyi/osproxy/app/pkg/repo"
	"github.com/qinguoyi/osproxy/app/pkg/storage"
	"github.com/qinguoyi/osproxy/app/pkg/utils"
	"github.com/qinguoyi/osproxy/app/pkg/web"
	"github.com/qinguoyi/osproxy/bootstrap/plugins"
	"go.uber.org/zap"
	"strconv"
	"time"
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

// GetMinioAdd 获取minio地址
//
//	@Summary		获取视频地址
//	@Description	获取视频地址
//	@Tags			列表
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	web.Response
//	@Router			/api/storage/v0/videos [get]
func GetMinioAdd(c *gin.Context) {
	uidStr := c.Query("uid")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		lgLogger.WithContext(c).Error("uid参数有误")
		web.ParamsError(c, err.Error())
		return
	}
	var meta *models.MetaDataInfo
	// 从redis缓存中查找
	lgRedis := new(plugins.LangGoRedis).NewRedis()
	val, err := lgRedis.Get(c, fmt.Sprintf("%s-meta", uidStr)).Result()
	if err == redis.Nil {
		// key在redis中不存在，从SQL中查找，再写入缓存
		lgDB := new(plugins.LangGoDB).Use("default").NewDB()
		meta, err = repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
		if err != nil {
			lgLogger.WithContext(c).Error(err.Error())
			lgLogger.WithContext(c).Error("下载数据，查询数据元信息失败", zap.String("traceId", err.Error()))
			web.InternalError(c, "内部异常")
			return
		}
		// 写入redis
		b, err := json.Marshal(meta)
		if err != nil {
			lgLogger.WithContext(c).Warn("下载数据，写入redis失败")
		}
		lgRedis.SetNX(c, fmt.Sprintf("%s-meta", uidStr), b, 5*60*time.Second)
	} else {
		if err != nil {
			lgLogger.WithContext(c).Error("下载数据，查询redis失败")
			web.InternalError(c, "")
			return
		}
		var msg models.MetaDataInfo
		if err := json.Unmarshal([]byte(val), &msg); err != nil {
			lgLogger.WithContext(c).Error("下载数据，查询redis结果，序列化失败")
			web.InternalError(c, "")
			return
		}
		// 续期
		lgRedis.Expire(c, fmt.Sprintf("%s-meta", uidStr), 5*60*time.Second)
		meta = &msg
	}
	bucketName := meta.Bucket
	objectName := meta.StorageName
	url, err := storage.NewStorage().Storage.GetFileUrl(bucketName, objectName)
	if err != nil {
		panic(err)
	}
	web.Success(c, url)
}
