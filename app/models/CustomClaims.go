package models

import "github.com/dgrijalva/jwt-go"

type CustomClaims struct {
	ID                 int64
	NickName           string
	AuthorityId        uint
	jwt.StandardClaims // 授权者、主题、发布时间、过期时间等
}
