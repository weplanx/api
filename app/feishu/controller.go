package feishu

import (
	"api/app/system"
	"api/app/users"
	"api/common"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	jsoniter "github.com/json-iterator/go"
	"github.com/weplanx/go/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

type Controller struct {
	*common.Inject
	Feishu *Service
	System *system.Service
	Users  *users.Service
}

func (x *Controller) Challenge(c *gin.Context) interface{} {
	var body struct {
		Encrypt string `json:"encrypt"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	option, err := x.System.GetVars(ctx, []string{"feishu_encrypt_key", "feishu_verification_token"})
	if err != nil {
		return err
	}
	content, err := x.Feishu.Decrypt(body.Encrypt, option["feishu_encrypt_key"].(string))
	if err != nil {
		return err
	}
	var dto struct {
		Challenge string `json:"challenge"`
		Token     string `json:"token"`
		Type      string `json:"type"`
	}
	if err = jsoniter.Unmarshal([]byte(content), &dto); err != nil {
		return err
	}
	if dto.Token != option["feishu_verification_token"] {
		return errors.New("验证令牌不一致")
	}
	return gin.H{
		"challenge": dto.Challenge,
	}
}

type State struct {
	Action string `json:"action,omitempty"`
}

func (x *Controller) OAuth(c *gin.Context) interface{} {
	var query struct {
		Code  string `form:"code" binding:"required"`
		State string `form:"state"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	ctx := c.Request.Context()
	userData, err := x.Feishu.GetAccessToken(ctx, query.Code)
	if err != nil {
		return err
	}
	var state State
	if err = jsoniter.Unmarshal([]byte(query.State), &state); err != nil {
		return err
	}
	switch state.Action {
	case "link":
		ts, err := c.Cookie("access_token")
		if err != nil {
			c.Set("status_code", 401)
			c.Set("code", "AUTH_EXPIRED")
			return common.AuthExpired
		}
		var claims jwt.MapClaims
		if claims, err = x.Passport.Verify(ts); err != nil {
			c.Set("status_code", 401)
			c.Set("code", "AUTH_EXPIRED")
			return err
		}
		userId, _ := primitive.ObjectIDFromHex(claims["context"].(map[string]interface{})["uid"].(string))
		if err := x.Users.UpdateOneById(ctx, userId, bson.M{
			"$set": bson.M{
				"feishu": userData,
			},
		}); err != nil {
			return err
		}
		c.Redirect(302, "https://xconsole.kainonly.com:8443/#/authorized")
		return nil
	}
	data, err := x.Users.FindOneByFeishu(ctx, userData.OpenId)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.Redirect(302, "https://xconsole.kainonly.com:8443/#/unauthorize")
		}
		return err
	}

	// 创建 Token
	jti := helper.Uuid()
	var ts string
	if ts, err = x.Passport.Create(jti, gin.H{
		"uid": data.ID.Hex(),
	}); err != nil {
		return err
	}
	// 设置会话
	if err := x.System.SetSession(ctx, data.ID.Hex(), jti); err != nil {
		return err
	}
	// 写入日志
	logData := common.NewLoginLogV10(data, jti, c.ClientIP(), c.Request.UserAgent())
	go x.System.PushLoginLog(context.TODO(), logData)
	// 返回
	c.SetCookie("access_token", ts, 0, "", "", true, true)
	c.SetSameSite(http.SameSiteStrictMode)
	c.Redirect(302, "https://xconsole.kainonly.com:8443/")
	return nil
}

// Option 获取配置
func (x *Controller) Option(c *gin.Context) interface{} {
	ctx := c.Request.Context()
	option, err := x.System.GetVars(ctx, []string{"redirect_url", "feishu_app_id"})
	if err != nil {
		return err
	}
	return gin.H{
		"url":      "https://open.feishu.cn/open-apis/authen/v1/index",
		"redirect": option["redirect_url"],
		"app_id":   option["feishu_app_id"],
	}
}
