package api

import (
	"github.com/gin-gonic/gin"
	v0 "github.com/qinguoyi/osproxy/api/v0"
	"github.com/qinguoyi/osproxy/app/middleware"
	"github.com/qinguoyi/osproxy/bootstrap"
	"github.com/qinguoyi/osproxy/config"
	"github.com/qinguoyi/osproxy/docs"
	gs "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func NewRouter(
	conf *config.Configuration,
	lgLogger *bootstrap.LangGoLogger,
) *gin.Engine {
	if conf.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// middleware
	corsM := middleware.NewCors()
	traceL := middleware.NewTrace(lgLogger)
	requestL := middleware.NewRequestLog(lgLogger)
	panicRecover := middleware.NewPanicRecover(lgLogger)

	// 跨域 trace-id 日志
	router.Use(corsM.Handler(), traceL.Handler(), requestL.Handler(), panicRecover.Handler())
	//router.Use(corsM.Handler(), traceL.Handler(), requestL.Handler())

	// 静态资源
	router.StaticFile("/assets", "../../static/image/back.png")

	// swag docs
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", gs.WrapHandler(swaggerFiles.Handler))

	// 动态资源 注册 api 分组路由
	setApiGroupRoutes(router)

	return router
}

func setApiGroupRoutes(
	router *gin.Engine,
) *gin.RouterGroup {
	jwt := middleware.NewJWT()
	group := router.Group("/api/storage/v0")
	{
		//health
		group.GET("/ping", v0.PingHandler)
		group.GET("/health", v0.HealthCheckHandler)

		// resume
		group.POST("/resume", v0.ResumeHandler)
		group.GET("/checkpoint", v0.CheckPointHandler)

		// link
		group.POST("/link/upload", v0.UploadLinkHandler)
		group.POST("/link/download", v0.DownloadLinkHandler)

		// proxy
		group.GET("/proxy", v0.IsOnCurrentServerHandler)

		//download
		group.GET("/download", v0.DownloadHandler1)
		//获取文件列表
		group.GET("/fileList", v0.GetFileList)
		// 视频播放Uid
		group.GET("/videos", v0.Videos)
		// 注册
		group.POST("register", v0.SignUp)
		//登录
		group.POST("/login", v0.LoginHandler)
		//
		group.GET("/showComment", v0.ShowCommentHandler)
		// 直接桶播放 获取minio视频流
		group.GET("/minioStream", v0.GetMinioAdd)
		//upload
		group.Use(jwt.JWTAuth())
		{
			// 文件上传
			group.PUT("/supload", v0.UploadSingleHandler1)
			group.PUT("/mupload1", v0.UploadHandler1)
			group.PUT("/mupload2", v0.UploadHandler2)
			group.PUT("/mupload3", v0.UploadHandler3)
			//用户评论
			group.POST("/sendComment", v0.SendCommentHandler)
		}
		//group.Use(jwt.JWTAuth()).PUT("/supload", v0.UploadSingleHandler1)
		//group.Use(jwt.JWTAuth()).PUT("/mupload1", v0.UploadHandler1)
		//group.Use(jwt.JWTAuth()).PUT("/mupload2", v0.UploadHandler2)
		//group.Use(jwt.JWTAuth()).PUT("/mupload3", v0.UploadHandler3)
		//
		//group.PUT("/upload", v0.UploadSingleHandler)
		//group.PUT("/upload/multi", v0.UploadMultiPartHandler)
		//group.PUT("/upload/merge", v0.UploadMergeHandler)

	}
	return group
}
