package v0

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/qinguoyi/osproxy/app/models"
	"github.com/qinguoyi/osproxy/app/pkg/base"
	"github.com/qinguoyi/osproxy/app/pkg/repo"
	"github.com/qinguoyi/osproxy/app/pkg/web"
	"github.com/qinguoyi/osproxy/bootstrap/plugins"
	"strconv"
	time2 "time"
)

// SendCommentHandler .
// @Summary		发送评论
// @Description	发送评论
// @Tags			发送评论
// @Accept			application/json
// @Param			RequestBody	body	models.SendReq	true	"发送评论请求体"
// @Produce			application/json
// @Success			200	{object} 	web.Response
// @Router			/api/storage/v0/sendComment [post]
func SendCommentHandler(c *gin.Context) {
	req := models.SendReq{}
	if err := c.ShouldBindJSON(&req); err != nil {
		panic(err)
		web.ParamsError(c, "参数有误")
		return
	}
	// 判断传来参数
	// 待完成
	marshal, err2 := json.Marshal(req)
	if err2 != nil {
		panic(err2)
	}
	lgLogger.WithContext(c).Info(string(marshal))
	id, err := base.NewSnowFlake().NextId()
	if err != nil {
		panic(err)
	}
	time := time2.Now()
	userId, _ := c.Get("userId")
	i := userId.(int64)
	commentInfo := models.CommentInfo{
		UID:           id,
		Content:       req.Content,
		IsDelete:      false,
		UserId:        i,
		VideoId:       req.VideoId,
		LikeCount:     0,
		RootCommentId: req.RootCommentId,
		ToUserId:      req.ToUserId,
		CreateTime:    &time,
	}
	db := new(plugins.LangGoDB).Use("default").NewDB()
	err = repo.NewComRepo().Create(db, commentInfo)
	if err != nil {
		panic(err)
	}
	web.Success(c, "评论成功")
}

// ShowCommentHandler .
// @Summary		显示评论
// @Description	显示评论
// @Tags			显示评论
//
//	@Accept			application/json
//	@Param			uid			query	string	true	"视频uid"
//
// @Produce		application/json
// @Success		200	{object} web.Response
// @Router		/api/storage/v0/showComment [get]
func ShowCommentHandler(c *gin.Context) {
	value := c.Query("uid")
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		panic(err)
	}
	db := new(plugins.LangGoDB).Use("default").NewDB()
	req, err := repo.NewComRepo().SelectByUid(db, i)
	var resp []models.ShowComment
	for _, c := range req {
		var son []models.ShowComment
		somComment, err := repo.NewComRepo().SelectByParent(db, c.UID)
		if err != nil {
			panic(err)
		}
		for _, info := range somComment {
			comment := models.ShowComment{
				Content:       info.Content,
				UserId:        info.UserId,
				LikeCount:     info.LikeCount,
				RootCommentId: info.RootCommentId,
				ToUserId:      info.ToUserId,
				CreateTime:    info.CreateTime,
			}
			son = append(son, comment)
		}
		comment := models.ShowComment{
			Content:       c.Content,
			UserId:        c.UserId,
			LikeCount:     c.LikeCount,
			RootCommentId: c.RootCommentId,
			ToUserId:      c.ToUserId,
			CreateTime:    c.CreateTime,
			SonComment:    son,
		}
		resp = append(resp, comment)
	}
	web.Success(c, resp)
}
