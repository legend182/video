// Code generated by swaggo/swag. DO NOT EDIT.

package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "qinguoyi"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/storage/v0/checkpoint": {
            "get": {
                "description": "断点续传",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "断点续传"
                ],
                "summary": "断点续传",
                "parameters": [
                    {
                        "type": "string",
                        "description": "文件uid",
                        "name": "uid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/web.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "type": "integer"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/api/storage/v0/download": {
            "get": {
                "description": "下载数据",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "下载"
                ],
                "summary": "下载数据",
                "parameters": [
                    {
                        "type": "string",
                        "description": "文件uid",
                        "name": "uid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "文件名称",
                        "name": "name",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "是否在线",
                        "name": "online",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/fileList": {
            "get": {
                "description": "获取列表",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "列表"
                ],
                "summary": "获取列表",
                "parameters": [
                    {
                        "type": "string",
                        "description": "page",
                        "name": "page",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "pageSize",
                        "name": "pageSize",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/health": {
            "get": {
                "description": "健康检查",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "检查"
                ],
                "summary": "健康检查",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/link/download": {
            "post": {
                "description": "获取下载连接",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "链接"
                ],
                "summary": "获取下载连接",
                "parameters": [
                    {
                        "description": "下载链接请求体",
                        "name": "RequestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.GenDownload"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/web.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/models.GenDownloadResp"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/api/storage/v0/link/upload": {
            "post": {
                "description": "初始化上传连接",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "链接"
                ],
                "summary": "初始化上传连接",
                "parameters": [
                    {
                        "description": "生成上传链接请求体",
                        "name": "RequestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.GenUpload"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/web.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "$ref": "#/definitions/models.GenUploadResp"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/api/storage/v0/login": {
            "post": {
                "description": "用户登录",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "登录"
                ],
                "summary": "用户登录",
                "parameters": [
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "username",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "密码",
                        "name": "password",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/mupload1": {
            "put": {
                "description": "获取分片上传uid",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "上传"
                ],
                "summary": "获取分片上传uid",
                "parameters": [
                    {
                        "type": "string",
                        "description": "fileName",
                        "name": "fileName",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "totalChunks",
                        "name": "totalChunks",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "md5",
                        "name": "md5",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "fileSize",
                        "name": "fileSize",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/mupload2": {
            "put": {
                "description": "上传分片",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "上传"
                ],
                "summary": "上传分片",
                "parameters": [
                    {
                        "type": "file",
                        "description": "上传的文件",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "文件uid",
                        "name": "uid",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "md5",
                        "name": "md5",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "chunk",
                        "name": "chunkNum",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/mupload3": {
            "put": {
                "description": "合并分片",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "上传"
                ],
                "summary": "合并分片",
                "parameters": [
                    {
                        "type": "string",
                        "description": "文件uid",
                        "name": "uid",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/ping": {
            "get": {
                "description": "测试接口",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "测试"
                ],
                "summary": "测试接口",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/proxy": {
            "get": {
                "description": "询问文件是否在当前服务",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "proxy"
                ],
                "summary": "询问文件是否在当前服务",
                "parameters": [
                    {
                        "type": "string",
                        "description": "uid",
                        "name": "uid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/register": {
            "post": {
                "description": "注册用户",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "注册"
                ],
                "summary": "注册用户",
                "parameters": [
                    {
                        "type": "string",
                        "description": "昵称",
                        "name": "name",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "用户名",
                        "name": "username",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "密码",
                        "name": "password",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/resume": {
            "post": {
                "description": "秒传\u0026断点续传",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "秒传"
                ],
                "summary": "秒传\u0026断点续传",
                "parameters": [
                    {
                        "description": "秒传请求体",
                        "name": "RequestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.ResumeReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/web.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "data": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/models.ResumeResp"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/api/storage/v0/sendComment": {
            "post": {
                "description": "发送评论",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "发送评论"
                ],
                "summary": "发送评论",
                "parameters": [
                    {
                        "description": "发送评论请求体",
                        "name": "RequestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.SendReq"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/showComment": {
            "get": {
                "description": "显示评论",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "显示评论"
                ],
                "summary": "显示评论",
                "parameters": [
                    {
                        "type": "string",
                        "description": "视频uid",
                        "name": "uid",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/supload": {
            "put": {
                "description": "单文件上传",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "上传"
                ],
                "summary": "单文件上传",
                "parameters": [
                    {
                        "type": "file",
                        "description": "上传的文件",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/upload": {
            "put": {
                "description": "上传单个文件",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "上传"
                ],
                "summary": "上传单个文件",
                "parameters": [
                    {
                        "type": "file",
                        "description": "上传的文件",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "文件uid",
                        "name": "uid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "md5",
                        "name": "md5",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "链接生成时间",
                        "name": "date",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "过期时间",
                        "name": "expire",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "签名",
                        "name": "signature",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/upload/merge": {
            "put": {
                "description": "合并分片文件",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "上传"
                ],
                "summary": "合并分片文件",
                "parameters": [
                    {
                        "type": "string",
                        "description": "文件uid",
                        "name": "uid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "md5",
                        "name": "md5",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "总分片数量",
                        "name": "num",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "文件总大小",
                        "name": "size",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "链接生成时间",
                        "name": "date",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "过期时间",
                        "name": "expire",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "签名",
                        "name": "signature",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/upload/multi": {
            "put": {
                "description": "上传分片文件",
                "consumes": [
                    "multipart/form-data"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "上传"
                ],
                "summary": "上传分片文件",
                "parameters": [
                    {
                        "type": "file",
                        "description": "上传的文件",
                        "name": "file",
                        "in": "formData",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "文件uid",
                        "name": "uid",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "md5",
                        "name": "md5",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "当前分片id",
                        "name": "chunkNum",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "链接生成时间",
                        "name": "date",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "过期时间",
                        "name": "expire",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "签名",
                        "name": "signature",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        },
        "/api/storage/v0/videos": {
            "get": {
                "description": "获取视频",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "列表"
                ],
                "summary": "获取视频uid",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/web.Response"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "models.GenDownload": {
            "type": "object",
            "required": [
                "expire",
                "uid"
            ],
            "properties": {
                "expire": {
                    "description": "过期时间",
                    "type": "integer"
                },
                "uid": {
                    "description": "文件路径",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "models.GenDownloadResp": {
            "type": "object",
            "properties": {
                "meta": {
                    "$ref": "#/definitions/models.MetaInfo"
                },
                "uid": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "models.GenUpload": {
            "type": "object",
            "required": [
                "filePath"
            ],
            "properties": {
                "expire": {
                    "description": "过期时间",
                    "type": "integer"
                },
                "filePath": {
                    "description": "文件路径",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "models.GenUploadResp": {
            "type": "object",
            "properties": {
                "path": {
                    "type": "string"
                },
                "uid": {
                    "type": "string"
                },
                "url": {
                    "$ref": "#/definitions/models.UrlResult"
                }
            }
        },
        "models.MD5Name": {
            "type": "object",
            "properties": {
                "md5": {
                    "type": "string"
                },
                "path": {
                    "type": "string"
                }
            }
        },
        "models.MetaInfo": {
            "type": "object",
            "properties": {
                "dstName": {
                    "type": "string"
                },
                "height": {
                    "type": "integer"
                },
                "md5": {
                    "type": "string"
                },
                "size": {
                    "type": "string"
                },
                "srcName": {
                    "type": "string"
                },
                "width": {
                    "type": "integer"
                }
            }
        },
        "models.MultiUrlResult": {
            "type": "object",
            "properties": {
                "merge": {
                    "type": "string"
                },
                "upload": {
                    "type": "string"
                }
            }
        },
        "models.ResumeReq": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/models.MD5Name"
                    }
                }
            }
        },
        "models.ResumeResp": {
            "type": "object",
            "properties": {
                "md5": {
                    "type": "string"
                },
                "uid": {
                    "type": "string"
                }
            }
        },
        "models.SendReq": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "rootCommentId": {
                    "description": "UserId        int64  ` + "`" + `json:\"userId,string\"` + "`" + `",
                    "type": "string",
                    "example": "0"
                },
                "toUserId": {
                    "type": "string",
                    "example": "0"
                },
                "videoId": {
                    "type": "string",
                    "example": "0"
                }
            }
        },
        "models.UrlResult": {
            "type": "object",
            "properties": {
                "multi": {
                    "$ref": "#/definitions/models.MultiUrlResult"
                },
                "single": {
                    "type": "string"
                }
            }
        },
        "web.Response": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "自定义错误码",
                    "type": "integer"
                },
                "data": {
                    "description": "数据"
                },
                "message": {
                    "description": "信息",
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "127.0.0.1:8888",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "ObjectStorageProxy",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
