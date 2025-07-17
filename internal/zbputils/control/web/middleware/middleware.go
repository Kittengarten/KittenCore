// Package middleware 中间件
package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

const (
	Authorization             = `Authorization`
	AccessControlAllowOrigin  = `Access-Control-Allow-Origin`
	AccessControlAllowHeaders = `Access-Control-Allow-Headers`
	ContentLength             = `Content-Length`
)

const sep = `, `

const (
	code    = `code`
	result  = `result`
	message = `message`
	type_   = `type`
	success = `success`
	error_  = `error`
)

// LoginCache 登录缓存
var LoginCache = cache.New(24*time.Hour, 12*time.Hour)

// Cors 跨域
/**
 * @Description: 支持跨域访问
 * @return gin.HandlerFunc
 * example
 */
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get(`Origin`) // 请求头部
		if origin != `` {
			// 接收客户端发送的origin （重要！）
			c.Writer.Header().Set(AccessControlAllowOrigin, origin)
			// 服务器支持的所有跨域请求的方法
			c.Header(`Access-Control-Allow-Methods`,
				strings.Join([]string{
					http.MethodGet,
					http.MethodPost,
					http.MethodPut,
					http.MethodDelete,
					http.MethodOptions,
					`UPDATE`,
				}, sep))
			// 允许跨域设置可以返回其他子段，可以自定义字段
			c.Header(AccessControlAllowHeaders,
				strings.Join([]string{
					Authorization,
					ContentLength,
					`X-CSRF-Token`,
					`Token`,
					`Session`,
					`Content-Type`,
				}, sep))
			// 允许浏览器（客户端）可以解析的头部 （重要）
			c.Header(`Access-Control-Expose-Headers`,
				strings.Join([]string{
					ContentLength,
					AccessControlAllowOrigin,
					AccessControlAllowHeaders,
				}, sep))
			// 设置缓存时间
			c.Header(`Access-Control-Max-Age`, strconv.Itoa(172800))
			// 允许客户端传递校验信息比如 cookie (重要)
			c.Header(`Access-Control-Allow-Credentials`, strconv.FormatBool(true))
		}

		// 允许类型校验
		if method == http.MethodOptions {
			c.JSON(http.StatusOK, gin.H{
				code:    0,
				result:  nil,
				message: ``,
				type_:   success,
			})
		}

		c.Next()
	}
}

// TokenMiddle 验证token
func TokenMiddle() gin.HandlerFunc {
	return func(con *gin.Context) {
		// 进行token验证
		token := con.Request.Header.Get(Authorization)
		if token == `` {
			con.JSON(http.StatusUnauthorized, gin.H{
				code:    2,
				result:  nil,
				message: `无权访问, 请登录`,
				type_:   error_,
			})
			con.Abort()
			return
		}
		_, found := LoginCache.Get(token)
		if !found {
			con.JSON(http.StatusUnauthorized, gin.H{
				code:    2,
				result:  nil,
				message: `token 无效, 请重新登录`,
				type_:   error_,
			})
			con.Abort()
			return
		}
		con.Next()
	}
}
