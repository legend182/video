package v0

import (
	"crypto/sha512"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/qinguoyi/osproxy/app/middleware"
	"github.com/qinguoyi/osproxy/app/models"
	"github.com/qinguoyi/osproxy/app/pkg/base"
	"github.com/qinguoyi/osproxy/app/pkg/repo"
	"github.com/qinguoyi/osproxy/app/pkg/web"
	"github.com/qinguoyi/osproxy/bootstrap/plugins"
	"strings"
	"time"
)

// SignUp 注册用户
//
//	@Summary		注册用户
//	@Description	注册用户
//	@Tags			注册
//	@Accept			multipart/form-data
//	@Param			name			formData		string	true	"昵称"
//	@Param			username		formData		string	true	"用户名"
//	@Param			password		formData		string	true	"密码"
//	@Produce		application/json
//	@Success		200	{object}	web.Response
//	@Router			/api/storage/v0/register [post]
func SignUp(c *gin.Context) {
	//	Param			confirm_password	formData		string	true	"确认密码"
	var fo *models.RegisterForm
	if err := c.ShouldBind(&fo); err != nil {
		lgLogger.WithContext(c).Error("SignUp error")
		web.ParamsError(c, err.Error())
		return
	}
	db := new(plugins.LangGoDB).Use("default").NewDB()
	error := repo.NewUserRepo().CheckUserExist(db, fo.UserName)
	if error != nil {
		lgLogger.WithContext(c).Error("用户名已注册")
		web.ParamsError(c, error.Error())
		return
	}
	id, error := base.NewSnowFlake().NextId()
	options := &password.Options{SaltLen: 10, Iterations: 10000, KeyLen: 50, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode(fo.Password, options)
	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	user := models.User{
		UserID:   id,
		UserName: fo.UserName,
		Password: newPassword,
		Name:     fo.Name,
	}
	error = repo.NewUserRepo().Create(db, &user)
	if error != nil {
		lgLogger.WithContext(c).Error("注册失败")
		web.ParamsError(c, error.Error())
		return
	}
	web.Success(c, "注册成功")
}

// LoginHandler 用户登录
//
//	@Summary		用户登录
//	@Description	用户登录
//	@Tags			登录
//	@Accept			multipart/form-data
//	@Param			username		formData		string	true	"用户名"
//	@Param			password		formData		string	true	"密码"
//	@Produce		application/json
//	@Success		200	{object}	web.Response
//	@Router			/api/storage/v0/login [post]
func LoginHandler(c *gin.Context) {
	var fo *models.LoginForm
	if err := c.ShouldBind(&fo); err != nil {
		lgLogger.WithContext(c).Error("Login error")
		web.ParamsError(c, err.Error())
		return
	}
	// 通过userName查询
	db := new(plugins.LangGoDB).Use("default").NewDB()
	user, err := repo.NewUserRepo().SelectByName(db, fo.UserName)
	if err != nil {
		panic(err)
	}
	split := strings.Split(user.Password, "$")
	options := &password.Options{SaltLen: 10, Iterations: 10000, KeyLen: 50, HashFunction: sha512.New}
	check := password.Verify(fo.Password, split[2], split[3], options)
	if check == false {
		lgLogger.WithContext(c).Error("密码错误")
		web.ParamsError(c, "密码错误")
	}
	claim := models.CustomClaims{
		ID:          user.UserID,
		NickName:    user.Name,
		AuthorityId: 0,
		StandardClaims: jwt.StandardClaims{ExpiresAt: time.Now().Add(
			time.Duration(8760) * time.Hour).Unix(), // 过期时间
			Issuer: "bluebell",
		},
	}
	token, err := middleware.NewJWT().CreateToken(claim)
	if err != nil {
		lgLogger.WithContext(c).Error("生成token出错")
	}
	web.SuccessLogin(c, token)
}
