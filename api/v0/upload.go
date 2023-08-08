package v0

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/qinguoyi/osproxy/app/models"
	"github.com/qinguoyi/osproxy/app/pkg/base"
	"github.com/qinguoyi/osproxy/app/pkg/repo"
	"github.com/qinguoyi/osproxy/app/pkg/storage"
	"github.com/qinguoyi/osproxy/app/pkg/thirdparty"
	"github.com/qinguoyi/osproxy/app/pkg/utils"
	"github.com/qinguoyi/osproxy/app/pkg/web"
	"github.com/qinguoyi/osproxy/bootstrap"
	"github.com/qinguoyi/osproxy/bootstrap/plugins"
	"go.uber.org/zap"
	"io"
	"os"
	"path"
	"strconv"
	"sync"
	"time"
)

/*
对象上传
*/

// UploadSingleHandler    上传单个文件
//
//	@Summary		上传单个文件
//	@Description	上传单个文件
//	@Tags			上传
//	@Accept			multipart/form-data
//	@Param			file		formData	file	true	"上传的文件"
//	@Param			uid			query		string	true	"文件uid"
//	@Param			md5			query		string	true	"md5"
//	@Param			date		query		string	true	"链接生成时间"
//	@Param			expire		query		string	true	"过期时间"
//	@Param			signature	query		string	true	"签名"
//	@Produce		application/json
//	@Success		200	{object}	web.Response
//	@Router			/api/storage/v0/upload [put]
func UploadSingleHandler(c *gin.Context) {
	uidStr := c.Query("uid")
	md5 := c.Query("md5")
	date := c.Query("date")
	expireStr := c.Query("expire")
	signature := c.Query("signature")

	uid, err, errorInfo := base.CheckValid(uidStr, date, expireStr)
	if err != nil {
		web.ParamsError(c, errorInfo)
		return
	}

	if !base.CheckUploadSignature(date, expireStr, signature) {
		web.ParamsError(c, "签名校验失败")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		web.ParamsError(c, fmt.Sprintf("解析文件参数失败，详情：%s", err))
		return
	}

	// 判断记录是否存在
	lgDB := new(plugins.LangGoDB).Use("default").NewDB()
	metaData, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
	if err != nil {
		web.NotFoundResource(c, "当前上传链接无效，uid不存在")
		return
	}

	dirName := path.Join(utils.LocalStore, uidStr)
	// 判断是否上传过，md5
	resumeInfo, err := repo.NewMetaDataInfoRepo().GetResumeByMd5(lgDB, []string{md5})
	if err != nil {
		lgLogger.WithContext(c).Error("查询文件是否已上传失败")
		web.InternalError(c, "")
		return
	}
	if len(resumeInfo) != 0 {
		now := time.Now()
		if err := repo.NewMetaDataInfoRepo().Updates(lgDB, uid, map[string]interface{}{
			"bucket":       resumeInfo[0].Bucket,
			"storage_name": resumeInfo[0].StorageName,
			"address":      resumeInfo[0].Address,
			"md5":          md5,
			"storage_size": resumeInfo[0].StorageSize,
			"multi_part":   false,
			"status":       1,
			"updated_at":   &now,
			"content_type": resumeInfo[0].ContentType,
		}); err != nil {
			lgLogger.WithContext(c).Error("上传完更新数据失败")
			web.InternalError(c, "上传完更新数据失败")
			return
		}
		// 可能有错误，如果目录不再本地也是可以秒传的，需要向其他服务器发送请求删除目录
		if err := os.RemoveAll(dirName); err != nil {
			lgLogger.WithContext(c).Error(fmt.Sprintf("删除目录失败，详情%s", err.Error()))
			web.InternalError(c, fmt.Sprintf("删除目录失败，详情%s", err.Error()))
			return
		}
		// 首次写入redis 元数据
		lgRedis := new(plugins.LangGoRedis).NewRedis()
		metaCache, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
		if err != nil {
			lgLogger.WithContext(c).Error("上传数据，查询数据元信息失败")
			web.InternalError(c, "内部异常")
			return
		}
		b, err := json.Marshal(metaCache)
		if err != nil {
			lgLogger.WithContext(c).Warn("上传数据，写入redis失败")
		}
		lgRedis.SetNX(context.Background(), fmt.Sprintf("%s-meta", uidStr), b, 5*60*time.Second)

		web.Success(c, "")
		return
	}
	// 判断是否在本地
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		// 不在本地，询问集群内其他服务并转发
		serviceList, err := base.NewServiceRegister().Discovery()
		if err != nil || serviceList == nil {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		var wg sync.WaitGroup
		var ipList []string
		ipChan := make(chan string, len(serviceList))
		for _, service := range serviceList {
			wg.Add(1)
			go func(ip string, port string, ipChan chan string, wg *sync.WaitGroup) {
				defer wg.Done()
				res, err := thirdparty.NewStorageService().Locate(utils.Scheme, ip, port, uidStr)
				if err != nil {
					fmt.Print(err.Error())
					return
				}
				ipChan <- res
			}(service.IP, service.Port, ipChan, &wg)
		}
		wg.Wait()
		close(ipChan)
		for re := range ipChan {
			ipList = append(ipList, re)
		}
		if len(ipList) == 0 {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		proxyIP := ipList[0]

		_, _, _, err = thirdparty.NewStorageService().UploadForward(c, utils.Scheme, proxyIP,
			bootstrap.NewConfig("").App.Port, uidStr, true)
		if err != nil {
			lgLogger.WithContext(c).Error("上传单文件，转发失败")
			web.InternalError(c, err.Error())
			return
		}
		web.Success(c, "")
		return
	}
	// 在本地
	fileName := path.Join(utils.LocalStore, uidStr, metaData.StorageName)
	out, err := os.Create(fileName)
	if err != nil {
		lgLogger.WithContext(c).Error("本地创建文件失败")
		web.InternalError(c, "本地创建文件失败")
		return
	}
	src, err := file.Open()
	if err != nil {
		lgLogger.WithContext(c).Error("打开本地文件失败")
		web.InternalError(c, "打开本地文件失败")
		return
	}
	if _, err = io.Copy(out, src); err != nil {
		lgLogger.WithContext(c).Error("请求数据存储到文件失败")
		web.InternalError(c, "请求数据存储到文件失败")
		return
	}
	// 校验md5
	md5Str, err := base.CalculateFileMd5(fileName)
	if err != nil {
		lgLogger.WithContext(c).Error(fmt.Sprintf("生成md5失败，详情%s", err.Error()))
		web.InternalError(c, err.Error())
		return
	}
	if md5Str != md5 {
		web.ParamsError(c, fmt.Sprintf("校验md5失败，计算结果:%s, 参数:%s", md5Str, md5))
		return
	}
	// 上传到minio
	contentType, err := base.DetectContentType(fileName)
	if err != nil {
		lgLogger.WithContext(c).Error("判断文件content-type失败")
		web.InternalError(c, "判断文件content-type失败")
		return
	}
	if err := storage.NewStorage().Storage.PutObject(metaData.Bucket, metaData.StorageName, fileName, contentType); err != nil {
		lgLogger.WithContext(c).Error("上传到minio失败")
		web.InternalError(c, "上传到minio失败")
		return
	}
	// 更新元数据
	now := time.Now()
	fileInfo, _ := os.Stat(fileName)
	if err := repo.NewMetaDataInfoRepo().Updates(lgDB, metaData.UID, map[string]interface{}{
		"md5":          md5Str,
		"storage_size": fileInfo.Size(),
		"multi_part":   false,
		"status":       1,
		"updated_at":   &now,
		"content_type": contentType,
	}); err != nil {
		lgLogger.WithContext(c).Error("上传完更新数据失败")
		web.InternalError(c, "上传完更新数据失败")
		return
	}
	_, _ = out.Close(), src.Close()

	if err := os.RemoveAll(dirName); err != nil {
		lgLogger.WithContext(c).Error(fmt.Sprintf("删除目录失败，详情%s", err.Error()))
		web.InternalError(c, fmt.Sprintf("删除目录失败，详情%s", err.Error()))
		return
	}

	// 首次写入redis 元数据
	lgRedis := new(plugins.LangGoRedis).NewRedis()
	metaCache, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
	if err != nil {
		lgLogger.WithContext(c).Error("上传数据，查询数据元信息失败")
		web.InternalError(c, "内部异常")
		return
	}
	b, err := json.Marshal(metaCache)
	if err != nil {
		lgLogger.WithContext(c).Warn("上传数据，写入redis失败")
	}
	lgRedis.SetNX(context.Background(), fmt.Sprintf("%s-meta", uidStr), b, 5*60*time.Second)

	web.Success(c, "")
	return
}

// UploadMultiPartHandler    上传分片文件
//
//	@Summary		上传分片文件
//	@Description	上传分片文件
//	@Tags			上传
//	@Accept			multipart/form-data
//	@Param			file		formData	file	true	"上传的文件"
//	@Param			uid			query		string	true	"文件uid"
//	@Param			md5			query		string	true	"md5"
//	@Param			chunkNum	query		string	true	"当前分片id"
//	@Param			date		query		string	true	"链接生成时间"
//	@Param			expire		query		string	true	"过期时间"
//	@Param			signature	query		string	true	"签名"
//	@Produce		application/json
//	@Success		200	{object}	web.Response
//	@Router			/api/storage/v0/upload/multi [put]
func UploadMultiPartHandler(c *gin.Context) {
	uidStr := c.Query("uid")
	md5 := c.Query("md5")
	// 与单文件上传相比少了多了一个文件序号
	chunkNumStr := c.Query("chunkNum")
	date := c.Query("date")
	expireStr := c.Query("expire")
	signature := c.Query("signature")

	uid, err, errorInfo := base.CheckValid(uidStr, date, expireStr)
	if err != nil {
		web.ParamsError(c, errorInfo)
		return
	}

	chunkNum, err := strconv.ParseInt(chunkNumStr, 10, 64)
	if err != nil {
		web.ParamsError(c, err.Error())
		return
	}

	if !base.CheckUploadSignature(date, expireStr, signature) {
		web.ParamsError(c, "签名校验失败")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		web.ParamsError(c, fmt.Sprintf("解析文件参数失败，详情：%s", err))
		return
	}

	// 判断记录是否存在
	lgDB := new(plugins.LangGoDB).Use("default").NewDB()
	metaData, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
	if err != nil {
		web.NotFoundResource(c, "当前上传链接无效，uid不存在")
		return
	}
	// 判断当前分片是否已上传??  逻辑有问题
	var lgRedis = new(plugins.LangGoRedis).NewRedis()
	ctx := context.Background()
	createLock := base.NewRedisLock(&ctx, lgRedis, fmt.Sprintf("multi-part-%d-%d-%s", uid, chunkNum, md5))
	if flag, err := createLock.Acquire(); err != nil || !flag {
		lgLogger.WithContext(c).Error("上传多文件抢锁失败")
		web.InternalError(c, "上传多文件抢锁失败")
		return
	}
	partInfo, err := repo.NewMultiPartInfoRepo().GetPartInfo(lgDB, uid, chunkNum, md5)
	if err != nil {
		lgLogger.WithContext(c).Error("多文件上传，查询分片信息失败")
		web.InternalError(c, "内部异常")
		return
	}
	if len(partInfo) != 0 {
		web.Success(c, "")
		return
	}
	_, _ = createLock.Release()

	// 判断是否在本地
	dirName := path.Join(utils.LocalStore, uidStr)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		// 不在本地，询问集群内其他服务并转发
		serviceList, err := base.NewServiceRegister().Discovery()
		if err != nil || serviceList == nil {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		var wg sync.WaitGroup
		var ipList []string
		ipChan := make(chan string, len(serviceList))
		for _, service := range serviceList {
			wg.Add(1)
			go func(ip string, port string, ipChan chan string, wg *sync.WaitGroup) {
				defer wg.Done()
				res, err := thirdparty.NewStorageService().Locate(utils.Scheme, ip, port, uidStr)
				if err != nil {
					fmt.Print(err.Error())
					return
				}
				ipChan <- res
			}(service.IP, service.Port, ipChan, &wg)
		}
		wg.Wait()
		close(ipChan)
		for re := range ipChan {
			ipList = append(ipList, re)
		}
		if len(ipList) == 0 {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		proxyIP := ipList[0]
		_, _, _, err = thirdparty.NewStorageService().UploadForward(c, utils.Scheme, proxyIP,
			bootstrap.NewConfig("").App.Port, uidStr, false)
		if err != nil {
			lgLogger.WithContext(c).Error("多文件上传，转发失败")
			web.InternalError(c, err.Error())
			return
		}
		web.Success(c, "")
		return
	}

	// 在本地
	fileName := path.Join(utils.LocalStore, uidStr, fmt.Sprintf("%d_%d", uid, chunkNum))
	out, err := os.Create(fileName)
	if err != nil {
		lgLogger.WithContext(c).Error("本地创建文件失败")
		web.InternalError(c, "本地创建文件失败")
		return
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)
	src, err := file.Open()
	if err != nil {
		lgLogger.WithContext(c).Error("打开本地文件失败")
		web.InternalError(c, "打开本地文件失败")
		return
	}
	if _, err = io.Copy(out, src); err != nil {
		lgLogger.WithContext(c).Error("请求数据存储到文件失败")
		web.InternalError(c, "请求数据存储到文件失败")
		return
	}
	// 校验md5
	md5Str, err := base.CalculateFileMd5(fileName)
	if err != nil {
		lgLogger.WithContext(c).Error(fmt.Sprintf("生成md5失败，详情%s", err.Error()))
		web.InternalError(c, err.Error())
		return
	}
	if md5Str != md5 {
		lgLogger.WithContext(c).Error(fmt.Sprintf("校验md5失败，计算结果:%s, 参数:%s", md5Str, md5))
		web.ParamsError(c, fmt.Sprintf("校验md5失败，计算结果:%s, 参数:%s", md5Str, md5))
		return
	}
	// 上传到minio
	contentType := "application/octet-stream"
	if err := storage.NewStorage().Storage.PutObject(metaData.Bucket, fmt.Sprintf("%d_%d", uid, chunkNum),
		fileName, contentType); err != nil {
		lgLogger.WithContext(c).Error("上传到minio失败")
		web.InternalError(c, "上传到minio失败")
		return
	}

	// 创建元数据
	now := time.Now()
	fileInfo, _ := os.Stat(fileName)
	if err := repo.NewMultiPartInfoRepo().Create(lgDB, &models.MultiPartInfo{
		StorageUid:   uid,
		ChunkNum:     int(chunkNum),
		Bucket:       metaData.Bucket,
		StorageName:  fmt.Sprintf("%d_%d", uid, chunkNum),
		StorageSize:  fileInfo.Size(),
		PartFileName: fmt.Sprintf("%d_%d", uid, chunkNum),
		PartMd5:      md5Str,
		Status:       1,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}); err != nil {
		lgLogger.WithContext(c).Error("上传完更新数据失败")
		web.InternalError(c, "上传完更新数据失败")
		return
	}
	web.Success(c, "")
	return
}

// UploadMergeHandler     合并分片文件
//
//	@Summary		合并分片文件
//	@Description	合并分片文件
//	@Tags			上传
//	@Accept			multipart/form-data
//	@Param			uid			query	string	true	"文件uid"
//	@Param			md5			query	string	true	"md5"
//	@Param			num			query	string	true	"总分片数量"
//	@Param			size		query	string	true	"文件总大小"
//	@Param			date		query	string	true	"链接生成时间"
//	@Param			expire		query	string	true	"过期时间"
//	@Param			signature	query	string	true	"签名"
//	@Produce		application/json
//	@Success		200	{object}	web.Response
//	@Router			/api/storage/v0/upload/merge [put]
func UploadMergeHandler(c *gin.Context) {
	uidStr := c.Query("uid")
	md5 := c.Query("md5")
	numStr := c.Query("num")
	size := c.Query("size")
	date := c.Query("date")
	expireStr := c.Query("expire")
	signature := c.Query("signature")

	uid, err, errorInfo := base.CheckValid(uidStr, date, expireStr)
	if err != nil {
		web.ParamsError(c, errorInfo)
		return
	}

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		web.ParamsError(c, errorInfo)
		return
	}

	if !base.CheckUploadSignature(date, expireStr, signature) {
		web.ParamsError(c, "签名校验失败")
		return
	}

	// 判断记录是否存在
	lgDB := new(plugins.LangGoDB).Use("default").NewDB()
	metaData, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
	if err != nil {
		web.NotFoundResource(c, "当前合并链接无效，uid不存在")
		return
	}

	// 判断分片数量是否一致
	var multiPartInfoList []models.MultiPartInfo
	if err := lgDB.Model(&models.MultiPartInfo{}).Where(
		"storage_uid = ? and status = ?", uid, 1).Order("chunk_num ASC").Find(&multiPartInfoList).Error; err != nil {
		lgLogger.WithContext(c).Error("查询分片数据失败")
		web.InternalError(c, "查询分片数据失败")
		return
	}

	if num != int64(len(multiPartInfoList)) {
		// 创建脏数据删除任务
		msg := models.MergeInfo{
			StorageUid: uid,
			ChunkSum:   num,
		}
		b, err := json.Marshal(msg)
		if err != nil {
			lgLogger.WithContext(c).Error("消息struct转成json字符串失败", zap.Any("err", err.Error()))
			web.InternalError(c, "分片数量和整体数量不一致，创建删除任务失败")
			return
		}
		newModelTask := models.TaskInfo{
			Status:    utils.TaskStatusUndo,
			TaskType:  utils.TaskPartDelete,
			ExtraData: string(b),
		}
		if err := repo.NewTaskRepo().Create(lgDB, &newModelTask); err != nil {
			lgLogger.WithContext(c).Error("分片数量和整体数量不一致，创建删除任务失败", zap.Any("err", err.Error()))
			web.InternalError(c, "分片数量和整体数量不一致，创建删除任务失败")
			return
		}
		web.ParamsError(c, "分片数量和整体数量不一致")
		return
	}

	// 判断是否在本地
	dirName := path.Join(utils.LocalStore, uidStr)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		// 不在本地，询问集群内其他服务并转发
		serviceList, err := base.NewServiceRegister().Discovery()
		if err != nil || serviceList == nil {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		var wg sync.WaitGroup
		var ipList []string
		ipChan := make(chan string, len(serviceList))
		for _, service := range serviceList {
			wg.Add(1)
			go func(ip string, port string, ipChan chan string, wg *sync.WaitGroup) {
				defer wg.Done()
				res, err := thirdparty.NewStorageService().Locate(utils.Scheme, ip, port, uidStr)
				if err != nil {
					return
				}
				ipChan <- res
			}(service.IP, service.Port, ipChan, &wg)
		}
		wg.Wait()
		close(ipChan)
		for re := range ipChan {
			ipList = append(ipList, re)
		}
		if len(ipList) == 0 {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		proxyIP := ipList[0]
		_, _, _, err = thirdparty.NewStorageService().MergeForward(c, utils.Scheme, proxyIP,
			bootstrap.NewConfig("").App.Port, uidStr)
		if err != nil {
			lgLogger.WithContext(c).Error("合并文件，转发失败")
			web.InternalError(c, err.Error())
			return
		}
		web.Success(c, "")
		return
	}
	// 获取文件的content-type
	firstPart := multiPartInfoList[0]
	partName := path.Join(utils.LocalStore, fmt.Sprintf("%d", uid), firstPart.PartFileName)
	contentType, err := base.DetectContentType(partName)
	if err != nil {
		lgLogger.WithContext(c).Error("判断文件content-type失败")
		web.InternalError(c, "判断文件content-type失败")
		return
	}
	// 更新metadata的数据
	now := time.Now()
	if err := repo.NewMetaDataInfoRepo().Updates(lgDB, metaData.UID, map[string]interface{}{
		"part_num":     int(num),
		"md5":          md5,
		"storage_size": size,
		"multi_part":   true,
		"status":       1,
		"updated_at":   &now,
		"content_type": contentType,
	}); err != nil {
		lgLogger.WithContext(c).Error("上传完更新数据失败")
		web.InternalError(c, "上传完更新数据失败")
		return
	}
	// 创建合并任务
	msg := models.MergeInfo{
		StorageUid: uid,
		ChunkSum:   num,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		lgLogger.WithContext(c).Error("消息struct转成json字符串失败", zap.Any("err", err.Error()))
		web.InternalError(c, "创建合并任务失败")
		return
	}
	newModelTask := models.TaskInfo{
		Status:    utils.TaskStatusUndo,
		TaskType:  utils.TaskPartMerge,
		ExtraData: string(b),
	}
	if err := repo.NewTaskRepo().Create(lgDB, &newModelTask); err != nil {
		lgLogger.WithContext(c).Error("创建合并任务失败", zap.Any("err", err.Error()))
		web.InternalError(c, "创建合并任务失败")
		return
	}

	// 首次写入redis 元数据和分片信息
	lgRedis := new(plugins.LangGoRedis).NewRedis()
	metaCache, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
	if err != nil {
		lgLogger.WithContext(c).Error("上传数据，查询数据元信息失败")
		web.InternalError(c, "内部异常")
		return
	}
	b, err = json.Marshal(metaCache)
	if err != nil {
		lgLogger.WithContext(c).Warn("上传数据，写入redis失败")
	}
	lgRedis.SetNX(context.Background(), fmt.Sprintf("%s-meta", uidStr), b, 5*60*time.Second)

	var multiPartInfoListCache []models.MultiPartInfo
	if err := lgDB.Model(&models.MultiPartInfo{}).Where(
		"storage_uid = ? and status = ?", uid, 1).Order("chunk_num ASC").Find(&multiPartInfoListCache).Error; err != nil {
		lgLogger.WithContext(c).Error("上传数据，查询分片数据失败")
		web.InternalError(c, "查询分片数据失败")
		return
	}
	// 写入redis
	b, err = json.Marshal(multiPartInfoListCache)
	if err != nil {
		lgLogger.WithContext(c).Warn("上传数据，写入redis失败")
	}
	lgRedis.SetNX(context.Background(), fmt.Sprintf("%s-multiPart", uidStr), b, 5*60*time.Second)

	web.Success(c, "")
	return
}

// UploadSingleHandler1   单文件上传
//
// @Summary		单文件上传
// @Description	单文件上传
// @Tags			上传
// @Accept			multipart/form-data
// @Produce			application/json
// @Param			file		formData	file	true	"上传的文件"
// @Success		200	{object}	web.Response
// @Router			/api/storage/v0/supload [put]
func UploadSingleHandler1(c *gin.Context) {
	// 获取文件属性
	file, err := c.FormFile("file")
	if err != nil {
		web.ParamsError(c, fmt.Sprintf("解析文件参数失败，详情：%s", err))
		return
	}
	// 为文件生成uid
	uid, err := base.NewSnowFlake().NextId()
	uidStr := strconv.FormatInt(uid, 10)
	//// 获取文件后缀
	//ext := base.GetExtension(file.Filename)
	// 存储文件名
	storageName := fmt.Sprintf("%s.%s", uidStr, base.GetExtension(file.Filename))
	// minio桶名称
	bucket := base.SelectBucketBySuffix(file.Filename)
	filePath := path.Join(utils.LocalStore, uidStr, storageName)
	// 文件在桶中地址
	objectName := fmt.Sprintf("%s/%s", bucket, storageName)
	os.MkdirAll(path.Join(utils.LocalStore, uidStr), 0666)
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		web.ParamsError(c, fmt.Sprintf("下载文件失败，详情：%s", err))
		return
	}
	// //通过md5判断文件是否存在
	// 将文件转为md5
	md5, err := base.CalculateFileMd5(filePath)
	if err != nil {
		web.ParamsError(c, fmt.Sprintf("解析文件失败，详情：%s", err))
		return
	}
	lgdb := new(plugins.LangGoDB).Use("default").NewDB()
	// 判断是否上传过
	resumeInfo, err := repo.NewMetaDataInfoRepo().GetResumeByMd5(lgdb, []string{md5})
	if err != nil {
		lgLogger.WithContext(c).Error("查询文件是否已上传失败")
		web.InternalError(c, "")
		return
	}
	// 已经上传直接就生成元数据
	if len(resumeInfo) != 0 {
		now := time.Now()
		if err := repo.NewMetaDataInfoRepo().Create(lgdb, &models.MetaDataInfo{
			UID:         uid,
			Bucket:      resumeInfo[0].StorageName,
			Name:        file.Filename,
			StorageName: resumeInfo[0].StorageName,
			Address:     resumeInfo[0].Address,
			Md5:         md5,
			StorageSize: file.Size,
			MultiPart:   false,
			Status:      1,
			ContentType: resumeInfo[0].ContentType,
			CreatedAt:   &now,
			UpdatedAt:   &now,
		}); err != nil {
			lgLogger.WithContext(c).Error("秒传时文件元数据生成失败")
			web.InternalError(c, "秒传时文件元数据生成失败")
			return
		}
		if err := os.RemoveAll(filePath); err != nil {
			lgLogger.WithContext(c).Error(fmt.Sprintf("删除目录失败，详情%s", err.Error()))
			web.InternalError(c, fmt.Sprintf("删除目录失败，详情%s", err.Error()))
			return
		}
		lgRedis := new(plugins.LangGoRedis).NewRedis()
		metaCache, err := repo.NewMetaDataInfoRepo().GetByUid(lgdb, uid)
		if err != nil {
			lgLogger.WithContext(c).Error("上传数据，查询数据元信息失败")
			web.InternalError(c, "内部异常")
			return
		}
		b, err := json.Marshal(metaCache)
		if err != nil {
			lgLogger.WithContext(c).Warn("上传数据，写入redis失败")
		}
		lgRedis.SetNX(context.Background(), fmt.Sprintf("%s-meta", uidStr), b, 5*60*time.Second)
		value, exists := c.Get("userId")
		if !exists {
			lgLogger.WithContext(c).Error("用户查找失败")
			web.InternalError(c, "用户查找失败,重新登录")
			return
		}
		i := value.(int64)
		err = repo.NewUFRepo().Create(lgdb, models.UserFile{
			UserId: i,
			FileId: uid,
		})
		if err != nil {
			panic(err)
			web.InternalError(c, "上传失败")
		}
		web.Success(c, "")
		return
	}
	// 没有上传过，上传到minio
	// 通过读文件头部内容获取文件的mime类型
	contentType, err := base.DetectContentType(filePath)
	if err != nil {
		lgLogger.WithContext(c).Error("判断文件content-type失败")
		web.InternalError(c, "判断文件content-type失败")
		return
	}
	//将文件上传到minio中
	if err := storage.NewStorage().Storage.PutObject(bucket, storageName, filePath, contentType); err != nil {
		lgLogger.WithContext(c).Error("上传到minio失败")
		web.InternalError(c, "上传到minio失败")
		return
	}

	now := time.Now()
	if err := repo.NewMetaDataInfoRepo().Create(lgdb, &models.MetaDataInfo{
		UID:         uid,
		Bucket:      bucket,
		Name:        file.Filename,
		StorageName: storageName,
		Address:     objectName,
		Md5:         md5,
		StorageSize: file.Size,
		MultiPart:   false,
		Status:      1,
		ContentType: contentType,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}); err != nil {
		lgLogger.WithContext(c).Error(err.Error())
		lgLogger.WithContext(c).Error(" 上传时文件元数据生成失败")
		web.InternalError(c, "上传时文件元数据生成失败")
		return
	}
	if err := os.RemoveAll(filePath); err != nil {
		lgLogger.WithContext(c).Error(fmt.Sprintf("删除目录失败，详情%s", err.Error()))
		web.InternalError(c, fmt.Sprintf("删除目录失败，详情%s", err.Error()))
		return
	}
	lgRedis := new(plugins.LangGoRedis).NewRedis()
	metaCache, err := repo.NewMetaDataInfoRepo().GetByUid(lgdb, uid)
	if err != nil {
		lgLogger.WithContext(c).Error("上传数据，查询数据元信息失败")
		web.InternalError(c, "内部异常")
		return
	}
	b, err := json.Marshal(metaCache)
	if err != nil {
		lgLogger.WithContext(c).Warn("上传数据，写入redis失败")
	}
	lgRedis.SetNX(context.Background(), fmt.Sprintf("%s-meta", uidStr), b, 5*60*time.Second)
	web.Success(c, "")
	return
}

// UploadHandler1   分片上传1
//
// @Summary		获取分片上传uid
// @Description	获取分片上传uid
// @Tags		上传
// @Accept		multipart/form-data
// @Produce		application/json
// @Param		fileName	    formData    string	true	"fileName"
// @Param		totalChunks		formData	string	true	"totalChunks"
// @Param		md5			    formData	string	true	"md5"
// @Param		fileSize		formData	string	true	"fileSize"
// @Success		200	{object}	web.Response
// @Router		/api/storage/v0/mupload1 [put]
func UploadHandler1(c *gin.Context) {
	uid, err := base.NewSnowFlake().NextId()
	uidStr := strconv.FormatInt(uid, 10)
	if err != nil {
		lgLogger.WithContext(c).Error("雪花算法生成ID失败，详情：", zap.Any("err", err.Error()))
		return
	}
	fileName := c.PostForm("fileName")
	fileSize := c.PostForm("fileSize")
	filesize, err := strconv.ParseInt(fileSize, 10, 64)
	if err != nil {
		lgLogger.WithContext(c).Error(err.Error())
		return
	}
	fmt.Println(fileName)
	if err != nil {
		lgLogger.WithContext(c).Error("文件名获取失败，详情：", zap.Any("err", err.Error()))
		return
	}
	md5 := c.PostForm("md5")
	if err != nil {
		lgLogger.WithContext(c).Error("文件名MD5失败，详情：", zap.Any("err", err.Error()))
		return
	}
	totalChunksStr := c.PostForm("totalChunks")
	fmt.Println(totalChunksStr)
	if err != nil {
		lgLogger.WithContext(c).Error("分片总数获取失败，详情：", zap.Any("err", err.Error()))
		return
	}
	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		lgLogger.WithContext(c).Error("分片总数转换失败，详情：", zap.Any("err", err.Error()))
		return
	}
	// 在本地创建uid的目录
	if err := os.MkdirAll(path.Join(utils.LocalStore, uidStr), 0755); err != nil {
		lgLogger.WithContext(c).Error("创建本地目录失败，详情：", zap.Any("err", err.Error()))
		return
	}
	now := time.Now()
	storageName := fmt.Sprintf("%s.%s", uidStr, base.GetExtension(fileName))
	// minio桶名称
	bucket := base.SelectBucketBySuffix(fileName)
	//filePath := path.Join(utils.LocalStore, uidStr, storageName)
	// 文件在桶中地址
	objectName := fmt.Sprintf("%s/%s", bucket, storageName)
	meta := models.MetaDataInfo{
		UID:         uid,
		Bucket:      bucket,
		Name:        fileName,
		StorageName: storageName,
		Address:     objectName,
		Md5:         md5,
		StorageSize: filesize,
		MultiPart:   true,
		PartNum:     totalChunks,
		Status:      -1,
		ContentType: "application/octet-stream", //先按照文件后缀占位，后面文件上传会覆盖
		CompressUid: 0,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}
	lgDB := new(plugins.LangGoDB).Use("default").NewDB()
	if err := repo.NewMetaDataInfoRepo().Create(lgDB, &meta); err != nil {
		lgLogger.WithContext(c).Error("生成链接数据库失败，详情：", zap.Any("err", err.Error()))
		web.InternalError(c, "内部异常")
		return
	}
	//  如果MD5存在且状态位为1就可以实现秒传了
	//
	web.Success(c, uidStr)
}

// UploadHandler2   分片上传2
//
// @Summary		上传分片
// @Description	上传分片
// @Tags			上传
// @Accept			multipart/form-data
// @Produce			application/json
// @Param			file		formData	file	true	"上传的文件"
// @Param			uid			formData	string	true	"文件uid"
// @Param			md5			formData	string	true	"md5"
// @Param			chunkNum		formData	string	true	"chunk"
// @Success			200	{object}	web.Response
// @Router			/api/storage/v0/mupload2 [put]
func UploadHandler2(c *gin.Context) {
	uidStr := c.PostForm("uid")
	md5 := c.PostForm("md5")
	// 与单文件上传相比少了多了一个文件序号
	chunkNumStr := c.PostForm("chunkNum")
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		lgLogger.WithContext(c).Info(err.Error())
		web.ParamsError(c, err.Error())
		return
	}
	chunkNum, err := strconv.ParseInt(chunkNumStr, 10, 64)
	if err != nil {
		lgLogger.WithContext(c).Info(err.Error())
		web.ParamsError(c, err.Error())
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		web.ParamsError(c, fmt.Sprintf("解析文件参数失败，详情：%s", err))
		return
	}
	// 判断记录是否存在,多文件上传系统会在某个服务器下生成文件并返回uid，所有分片应该传送个这个服务器
	lgDB := new(plugins.LangGoDB).Use("default").NewDB()
	metaData, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
	if err != nil {
		web.NotFoundResource(c, "当前上传链接无效，uid不存在")
		return
	}
	// 分片上传可能会重复发送分片，所以发送时应该上锁
	var lgRedis = new(plugins.LangGoRedis).NewRedis()
	ctx := context.Background()
	//md5 := c.PostForm("md5")
	createLock := base.NewRedisLock(&ctx, lgRedis, fmt.Sprintf("multi-part-%d-%d-%s", uid, chunkNum, md5))
	if flag, err := createLock.Acquire(); err != nil || !flag {
		lgLogger.WithContext(c).Error("上传多文件抢锁失败")
		web.InternalError(c, "上传多文件抢锁失败")
		return
	}

	// 判断该分片是否已经上传
	partInfo, err := repo.NewMultiPartInfoRepo().GetPartInfo(lgDB, uid, chunkNum, md5)
	if err != nil {
		lgLogger.WithContext(c).Error("多文件上传，查询分片信息失败")
		web.InternalError(c, "内部异常")
		return
	}
	if len(partInfo) != 0 {
		_, _ = createLock.Release()
		value, exists := c.Get("userId")
		if !exists {
			lgLogger.WithContext(c).Error("用户查找失败")
			web.InternalError(c, "用户查找失败,重新登录")
		}
		i := value.(int64)
		err = repo.NewUFRepo().Create(lgDB, models.UserFile{
			UserId: i,
			FileId: uid,
		})
		if err != nil {
			panic(err)
			web.InternalError(c, "上传失败")
		}
		web.Success(c, "")
		return
	}
	// 分片未上传情况
	//分片可以上传到任意服务器，但文件合并时应该在同一个服务器内，所以在生成uid时已经选择了
	//最终生成目标文件分服务器，判断当前分片是否在目标服务器内
	dirName := path.Join(utils.LocalStore, uidStr)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		//不在本地，向集群内转发，
		serviceList, err := base.NewServiceRegister().Discovery()
		// 没有其他节点或者发现服务失败
		if err != nil || serviceList == nil {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		// 询问存活的节点是否存在目录
		var wg sync.WaitGroup
		var ipList []string
		ipChan := make(chan string, len(serviceList))
		for _, service := range serviceList {
			wg.Add(1)
			go func(ip string, port string, ipChan chan string, wg *sync.WaitGroup) {
				defer wg.Done()
				//向目标ip发送文件是否存在的请求

				res, err := thirdparty.NewStorageService().Locate(utils.Scheme, ip, port, uidStr)
				if err != nil {
					fmt.Print(err.Error())
					return
				}
				ipChan <- res
			}(service.IP, service.Port, ipChan, &wg)
		}
		wg.Wait()
		close(ipChan)
		for re := range ipChan {
			ipList = append(ipList, re)
		}
		if len(ipList) == 0 {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		proxyIP := ipList[0]
		// 文件存在gin上下文的Body中，需要通过表单传递到目标地址
		_, _ = createLock.Release()
		_, _, _, err = thirdparty.NewStorageService().UploadForward(c, utils.Scheme, proxyIP,
			bootstrap.NewConfig("").App.Port, uidStr, false)
		if err != nil {
			lgLogger.WithContext(c).Error("多文件上传，转发失败")
			web.InternalError(c, err.Error())
			return
		}
		web.Success(c, "")
		return

	}
	// 在本地，分片上传将分片放入不同目录
	fileName := path.Join(dirName, fmt.Sprintf("%d_%d", uid, chunkNum))
	out, err := os.Create(fileName)
	if err != nil {
		lgLogger.WithContext(c).Error("本地创建文件失败")
		web.InternalError(c, "本地创建文件失败")
		return
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)
	src, err := file.Open()
	if err != nil {
		lgLogger.WithContext(c).Error("打开本地文件失败")
		web.InternalError(c, "打开本地文件失败")
		return
	}
	if _, err = io.Copy(out, src); err != nil {
		lgLogger.WithContext(c).Error("请求数据存储到文件失败")
		web.InternalError(c, "请求数据存储到文件失败")
		return
	}
	// 校验md5
	md5Str, err := base.CalculateFileMd5(fileName)
	if err != nil {
		lgLogger.WithContext(c).Error(fmt.Sprintf("生成md5失败，详情%s", err.Error()))
		web.InternalError(c, err.Error())
		return
	}
	if md5Str != md5 {
		lgLogger.WithContext(c).Error(fmt.Sprintf("校验md5失败，计算结果:%s, 参数:%s", md5Str, md5))
		web.ParamsError(c, fmt.Sprintf("校验md5失败，计算结果:%s, 参数:%s", md5Str, md5))
		return
	}
	// 上传到minio
	contentType := "application/octet-stream"
	if err := storage.NewStorage().Storage.PutObject(metaData.Bucket, fmt.Sprintf("%d_%d", uid, chunkNum),
		fileName, contentType); err != nil {
		lgLogger.WithContext(c).Error(err.Error())
		web.InternalError(c, "上传到minio失败")
		return
	}
	//创建元数据
	now := time.Now()
	fileInfo, _ := os.Stat(fileName)
	if err := repo.NewMultiPartInfoRepo().Create(lgDB, &models.MultiPartInfo{
		StorageUid:   uid,
		ChunkNum:     int(chunkNum),
		Bucket:       metaData.Bucket,
		StorageName:  fmt.Sprintf("%d_%d", uid, chunkNum),
		StorageSize:  fileInfo.Size(),
		PartFileName: fmt.Sprintf("%d_%d", uid, chunkNum),
		PartMd5:      md5Str,
		Status:       1,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}); err != nil {
		lgLogger.WithContext(c).Error("上传完更新数据失败")
		web.InternalError(c, "上传完更新数据失败")
		return
	}
	_, _ = createLock.Release()
	value, exists := c.Get("userId")
	if !exists {
		lgLogger.WithContext(c).Error("用户查找失败")
		web.InternalError(c, "用户查找失败,重新登录")
	}
	i := value.(int64)
	err = repo.NewUFRepo().Create(lgDB, models.UserFile{
		UserId: i,
		FileId: uid,
	})
	if err != nil {
		panic(err)
		web.InternalError(c, "上传失败")
	}
	web.Success(c, "")
	return
}

// UploadHandler3   合并
//
// @Summary		合并分片
// @Description	合并分片
// @Tags			上传
// @Accept			multipart/form-data
// @Produce			application/json
// @Param			uid				formData	string	true	"文件uid"
// @Success			200	{object}	web.Response
// @Router			/api/storage/v0/mupload3 [put]
func UploadHandler3(c *gin.Context) {
	uidStr := c.PostForm("uid")
	//md5 := c.Query("md5")
	//numStr := c.Query("num")
	//size := c.Query("size")
	//num, err := strconv.ParseInt(numStr, 10, 64)
	//if err != nil {
	//	web.ParamsError(c, err.Error())
	//	return
	//}
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		web.ParamsError(c, err.Error())
		return
	}
	//通过uid查询元数据判断记录是否存在
	lgDB := new(plugins.LangGoDB).Use("default").NewDB()
	metaData, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
	if err != nil {
		web.NotFoundResource(c, "当前合并链接无效，uid不存在")
		return
	}
	// 判断分片数量是否一致,将所有分片信息按照分片顺序递增查询
	var multiPartInfoList []models.MultiPartInfo
	if err := lgDB.Model(&models.MultiPartInfo{}).Where(
		"storage_uid = ? and status = ?", uid, 1).Order("chunk_num ASC").Find(&multiPartInfoList).Error; err != nil {
		lgLogger.WithContext(c).Error("查询分片数据失败")
		web.InternalError(c, "查询分片数据失败")
		return
	}
	//创建脏数据任务
	if metaData.PartNum != len(multiPartInfoList) {
		// 传递的分片数量和存储的分片数量不一样，将脏数据删除
		msg := models.MergeInfo{
			StorageUid: uid,
			ChunkSum:   int64(len(multiPartInfoList)),
		}
		b, err := json.Marshal(msg)
		if err != nil {
			lgLogger.WithContext(c).Error("消息struct转成json字符串失败", zap.Any("err", err.Error()))
			web.InternalError(c, "分片数量和整体数量不一致，创建删除任务失败")
			return
		}
		newModelTask := models.TaskInfo{
			Status:    utils.TaskStatusUndo,
			TaskType:  utils.TaskPartDelete,
			ExtraData: string(b),
		}
		if err := repo.NewTaskRepo().Create(lgDB, &newModelTask); err != nil {
			lgLogger.WithContext(c).Error("分片数量和整体数量不一致，创建删除任务失败", zap.Any("err", err.Error()))
			web.InternalError(c, "分片数量和整体数量不一致，创建删除任务失败")
			return
		}
		web.ParamsError(c, "分片数量和整体数量不一致")
		return
	}
	// 判断文件是否在本地
	dirName := path.Join(utils.LocalStore, uidStr)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		// 不在本地，询问集群内其他服务并转发
		serviceList, err := base.NewServiceRegister().Discovery()
		if err != nil || serviceList == nil {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		var wg sync.WaitGroup
		var ipList []string
		ipChan := make(chan string, len(serviceList))
		for _, service := range serviceList {
			wg.Add(1)
			go func(ip string, port string, ipChan chan string, wg *sync.WaitGroup) {
				defer wg.Done()
				res, err := thirdparty.NewStorageService().Locate(utils.Scheme, ip, port, uidStr)
				if err != nil {
					return
				}
				ipChan <- res
			}(service.IP, service.Port, ipChan, &wg)
		}
		wg.Wait()
		close(ipChan)
		for re := range ipChan {
			ipList = append(ipList, re)
		}
		if len(ipList) == 0 {
			lgLogger.WithContext(c).Error("发现其他服务失败")
			web.InternalError(c, "发现其他服务失败")
			return
		}
		proxyIP := ipList[0]
		// 转发文件的合并命令
		_, _, _, err = thirdparty.NewStorageService().MergeForward(c, utils.Scheme, proxyIP,
			bootstrap.NewConfig("").App.Port, uidStr)
		if err != nil {
			lgLogger.WithContext(c).Error("合并文件，转发失败")
			web.InternalError(c, err.Error())
			return
		}
		web.Success(c, "")
		return
	}
	//在本地
	// 获取文件的content-type
	firstPart := multiPartInfoList[0]
	partName := path.Join(utils.LocalStore, fmt.Sprintf("%d", uid), firstPart.PartFileName)
	contentType, err := base.DetectContentType(partName)
	if err != nil {
		lgLogger.WithContext(c).Error("判断文件content-type失败")
		web.InternalError(c, "判断文件content-type失败")
		return
	}
	// 更新metadata的数据
	now := time.Now()
	if err := repo.NewMetaDataInfoRepo().Updates(lgDB, metaData.UID, map[string]interface{}{
		"status":       1,
		"updated_at":   &now,
		"content_type": contentType,
	}); err != nil {
		lgLogger.WithContext(c).Error("上传完更新数据失败")
		web.InternalError(c, "上传完更新数据失败")
		return
	}
	// 创建合并任务
	msg := models.MergeInfo{
		StorageUid: uid,
		ChunkSum:   int64(metaData.PartNum),
	}
	b, err := json.Marshal(msg)
	if err != nil {
		lgLogger.WithContext(c).Error("消息struct转成json字符串失败", zap.Any("err", err.Error()))
		web.InternalError(c, "创建合并任务失败")
		return
	}
	newModelTask := models.TaskInfo{
		Status:    utils.TaskStatusUndo,
		TaskType:  utils.TaskPartMerge,
		ExtraData: string(b),
	}
	if err := repo.NewTaskRepo().Create(lgDB, &newModelTask); err != nil {
		lgLogger.WithContext(c).Error("创建合并任务失败", zap.Any("err", err.Error()))
		web.InternalError(c, "创建合并任务失败")
		return
	}
	// 首次写入redis 元数据和分片信息
	lgRedis := new(plugins.LangGoRedis).NewRedis()
	metaCache, err := repo.NewMetaDataInfoRepo().GetByUid(lgDB, uid)
	if err != nil {
		lgLogger.WithContext(c).Error("上传数据，查询数据元信息失败")
		web.InternalError(c, "内部异常")
		return
	}
	b, err = json.Marshal(metaCache)
	if err != nil {
		lgLogger.WithContext(c).Warn("上传数据，写入redis失败")
	}
	lgRedis.SetNX(context.Background(), fmt.Sprintf("%s-meta", uidStr), b, 5*60*time.Second)
	var multiPartInfoListCache []models.MultiPartInfo
	if err := lgDB.Model(&models.MultiPartInfo{}).Where(
		"storage_uid = ? and status = ?", uid, 1).Order("chunk_num ASC").Find(&multiPartInfoListCache).Error; err != nil {
		lgLogger.WithContext(c).Error("上传数据，查询分片数据失败")
		web.InternalError(c, "查询分片数据失败")
		return
	}
	// 写入redis
	b, err = json.Marshal(multiPartInfoListCache)
	if err != nil {
		lgLogger.WithContext(c).Warn("上传数据，写入redis失败")
	}
	lgRedis.SetNX(context.Background(), fmt.Sprintf("%s-multiPart", uidStr), b, 5*60*time.Second)
	// 写入redis缓存
	push := lgRedis.RPush(c, "video", metaData.UID)
	if push.Err() != nil {
		lgLogger.WithContext(c).Error(err.Error())
		panic(err)
	}
	web.Success(c, "")
	return
}
