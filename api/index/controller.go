package index

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/gookit/goutil/strutil"
	"github.com/weplanx/go/csrf"
	"github.com/weplanx/go/values"
	"github.com/weplanx/server/common"
	"github.com/weplanx/server/model"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"reflect"
	"time"
)

type Controller struct {
	V    *common.Values
	Csrf *csrf.Csrf

	IndexService  *Service
	ValuesService *values.Service
}

func (x *Controller) Ping(_ context.Context, c *app.RequestContext) {
	x.Csrf.SetToken(c)
	r := M{
		"name": x.V.Hostname,
		"ip":   string(c.GetHeader(x.V.Ip)),
		"now":  time.Now(),
	}
	if !x.V.IsRelease() {
		r["values"] = x.V
	}
	c.JSON(http.StatusOK, r)
}

type LoginDto struct {
	Email    string `json:"email,required" vd:"email($)"`
	Password string `json:"password,required" vd:"len($)>=8"`
}

func (x *Controller) Login(ctx context.Context, c *app.RequestContext) {
	var dto LoginDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	r, err := x.IndexService.Login(ctx, dto.Email, dto.Password)
	if err != nil {
		c.Error(err)
		return
	}

	go func() {
		data := model.NewLogsetLogined(
			r.User.ID, string(c.GetHeader(x.V.Ip)), "email", string(c.UserAgent()))
		if err = x.IndexService.WriteLogsetLogined(context.TODO(), data); err != nil {
			hlog.Fatal(err)
		}
	}()

	common.SetAccessToken(c, r.AccessToken)
	c.JSON(200, M{
		"code":    0,
		"message": "ok",
	})
}

type GetLoginSmsDto struct {
	Phone string `query:"phone,required"`
}

func (x *Controller) GetLoginSms(ctx context.Context, c *app.RequestContext) {
	var dto GetLoginSmsDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if _, err := x.IndexService.GetLoginSms(ctx, dto.Phone); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, M{
		"code":    0,
		"message": "ok",
	})
}

type LoginSmsDto struct {
	Phone string `json:"phone,required"`
	Code  string `json:"code,required"`
}

func (x *Controller) LoginSms(ctx context.Context, c *app.RequestContext) {
	var dto LoginSmsDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	r, err := x.IndexService.LoginSms(ctx, dto.Phone, dto.Code)
	if err != nil {
		c.Error(err)
		return
	}

	go func() {
		data := model.NewLogsetLogined(
			r.User.ID, string(c.GetHeader(x.V.Ip)), "sms", string(c.UserAgent()))
		if err = x.IndexService.WriteLogsetLogined(context.TODO(), data); err != nil {
			hlog.Fatal(err)
		}
	}()

	common.SetAccessToken(c, r.AccessToken)
	c.JSON(200, M{
		"code":    0,
		"message": "ok",
	})
}

type LoginTotpDto struct {
	Email string `json:"email,required" vd:"email($)"`
	Code  string `json:"code,required"`
}

func (x *Controller) LoginTotp(ctx context.Context, c *app.RequestContext) {
	var dto LoginTotpDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	r, err := x.IndexService.LoginTotp(ctx, dto.Email, dto.Code)
	if err != nil {
		c.Error(err)
		return
	}

	go func() {
		data := model.NewLogsetLogined(
			r.User.ID, string(c.GetHeader(x.V.Ip)), "totp", string(c.UserAgent()))
		if err = x.IndexService.WriteLogsetLogined(context.TODO(), data); err != nil {
			hlog.Fatal(err)
		}
	}()

	common.SetAccessToken(c, r.AccessToken)
	c.JSON(200, M{
		"code":    0,
		"message": "ok",
	})
}

type GetForgetCodeDto struct {
	Email string `query:"email,required" vd:"email($)"`
}

func (x *Controller) GetForgetCode(ctx context.Context, c *app.RequestContext) {
	var dto GetForgetCodeDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if err := x.IndexService.GetForgetCode(ctx, dto.Email); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, M{
		"code":    0,
		"message": "ok",
	})
}

type ForgetResetDto struct {
	Email    string `json:"email,required" vd:"email($)"`
	Code     string `json:"code,required"`
	Password string `json:"password,required"`
}

func (x *Controller) ForgetReset(ctx context.Context, c *app.RequestContext) {
	var dto ForgetResetDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if err := x.IndexService.ForgetReset(ctx, dto); err != nil {
		c.Error(err)
		return
	}

	c.JSON(200, M{
		"code":    0,
		"message": "ok",
	})
}

func (x *Controller) Verify(ctx context.Context, c *app.RequestContext) {
	ts := c.Cookie("TOKEN")
	if ts == nil {
		c.JSON(401, M{
			"code":    0,
			"message": common.ErrAuthenticationExpired.Error(),
		})
		return
	}

	if _, err := x.IndexService.Verify(ctx, string(ts)); err != nil {
		common.ClearAccessToken(c)
		c.JSON(401, M{
			"code":    0,
			"message": common.ErrAuthenticationExpired.Error(),
		})
		return
	}

	c.JSON(200, M{
		"code":    0,
		"message": "ok",
	})
}

func (x *Controller) GetRefreshCode(ctx context.Context, c *app.RequestContext) {
	claims := common.Claims(c)
	code, err := x.IndexService.GetRefreshCode(ctx, claims.UserId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, M{
		"code": code,
	})
}

type RefreshTokenDto struct {
	Code string `json:"code,required"`
}

func (x *Controller) RefreshToken(ctx context.Context, c *app.RequestContext) {
	var dto RefreshTokenDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	claims := common.Claims(c)
	ts, err := x.IndexService.RefreshToken(ctx, claims, dto.Code)
	if err != nil {
		c.Error(err)
		return
	}

	common.SetAccessToken(c, ts)
	c.JSON(http.StatusOK, M{
		"code":    0,
		"message": "ok",
	})
}

func (x *Controller) Logout(ctx context.Context, c *app.RequestContext) {
	claims := common.Claims(c)
	x.IndexService.Logout(ctx, claims.UserId)
	common.ClearAccessToken(c)
	c.JSON(http.StatusOK, M{
		"code":    0,
		"message": "ok",
	})
}

func (x *Controller) GetUser(ctx context.Context, c *app.RequestContext) {
	claims := common.Claims(c)
	data, err := x.IndexService.GetUser(ctx, claims.UserId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, data)
}

type SetUserDto struct {
	Key    string `json:"key,required" vd:"in($, 'email', 'name', 'avatar')"`
	Email  string `json:"email,omitempty" vd:"(Set)$!='Email' || email($);msg:'must be email'"`
	Name   string `json:"name,omitempty"`
	Avatar string `json:"avatar,omitempty"`
}

func (x *Controller) SetUser(ctx context.Context, c *app.RequestContext) {
	var dto SetUserDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	data := make(map[string]interface{})
	path := strutil.CamelCase(dto.Key)
	data[dto.Key] = reflect.ValueOf(dto).FieldByName(path).Interface()

	claims := common.Claims(c)
	if _, err := x.IndexService.SetUser(ctx, claims.UserId, bson.M{"$set": data}); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, M{
		"code":    0,
		"message": "ok",
	})
}

type SetUserPassword struct {
	Old      string `json:"old,required" vd:"len($)>8;msg:'must be greater than 8 characters'"`
	Password string `json:"password,required" vd:"len($)>8;msg:'must be greater than 8 characters'"`
}

func (x *Controller) SetUserPassword(ctx context.Context, c *app.RequestContext) {
	var dto SetUserPassword
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	claims := common.Claims(c)
	if _, err := x.IndexService.SetUserPassword(ctx, claims.UserId, dto.Old, dto.Password); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, M{
		"code":    0,
		"message": "ok",
	})
}

type GetUserPhone struct {
	Phone string `query:"phone,required"`
}

func (x *Controller) GetUserPhoneCode(ctx context.Context, c *app.RequestContext) {
	var dto GetUserPhone
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	if _, err := x.IndexService.GetUserPhoneCode(ctx, dto.Phone); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, M{
		"code":    0,
		"message": "ok",
	})
}

type SetUserPhone struct {
	Phone string `json:"phone,required"`
	Code  string `json:"code,required"`
}

func (x *Controller) SetUserPhone(ctx context.Context, c *app.RequestContext) {
	var dto SetUserPhone
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	claims := common.Claims(c)
	if _, err := x.IndexService.SetUserPhone(ctx, claims.UserId, dto.Phone, dto.Code); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, M{
		"code":    0,
		"message": "ok",
	})
}

func (x *Controller) GetUserTotp(ctx context.Context, c *app.RequestContext) {
	claims := common.Claims(c)
	r, err := x.IndexService.GetUserTotp(ctx, claims.UserId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, M{
		"totp": r,
	})
}

type SetUserTotp struct {
	Totp string    `json:"totp,required"`
	Tss  [2]string `json:"tss,required"`
}

func (x *Controller) SetUserTotp(ctx context.Context, c *app.RequestContext) {
	var dto SetUserTotp
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	claims := common.Claims(c)
	if _, err := x.IndexService.SetUserTotp(ctx, claims.UserId, dto.Totp, dto.Tss); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, M{
		"code":    0,
		"message": "ok",
	})
}

type UnsetUserDto struct {
	Key string `path:"key,required" vd:"in($, 'phone', 'totp', 'lark')"`
}

func (x *Controller) UnsetUser(ctx context.Context, c *app.RequestContext) {
	var dto UnsetUserDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}

	claims := common.Claims(c)
	_, err := x.IndexService.SetUser(ctx, claims.UserId, bson.M{
		"$unset": bson.M{dto.Key: 1},
	})
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, M{
		"code":    0,
		"message": "ok",
	})
}

type OptionsDto struct {
	Type string `query:"type"`
}

func (x *Controller) Options(ctx context.Context, c *app.RequestContext) {
	var dto OptionsDto
	if err := c.BindAndValidate(&dto); err != nil {
		c.Error(err)
		return
	}
	switch dto.Type {
	case "upload":
		switch x.V.Cloud {
		case "tencent":
			c.JSON(http.StatusOK, M{
				"type": "cos",
				"url": fmt.Sprintf(`https://%s.cos.%s.myqcloud.com`,
					x.V.TencentCosBucket, x.V.TencentCosRegion,
				),
				"limit": x.V.TencentCosLimit,
			})
			return
		}
	case "collaboration":
		c.JSON(http.StatusOK, M{
			"url":      "https://open.feishu.cn/open-apis/authen/v1/index",
			"redirect": x.V.RedirectUrl,
			"app_id":   x.V.LarkAppId,
		})
		return
	case "generate-secret":
		id, _ := strutil.RandomString(8)
		key, _ := strutil.RandomString(16)
		c.JSON(http.StatusOK, M{
			"id":  id,
			"key": key,
		})
		return
	}
	return
}
