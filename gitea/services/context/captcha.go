// Copyright 2020 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package context

import (
	"fmt"

	"github.com/wekan/caddy-gitea-plugin/gitea/modules/hcaptcha"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/imagecaptcha"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/log"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/mcaptcha"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/recaptcha"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/setting"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/templates"
	"github.com/wekan/caddy-gitea-plugin/gitea/modules/turnstile"
)

const (
	imageCaptchaIDField      = "captcha_id"
	imageCaptchaAnswerField  = "captcha"
	gRecaptchaResponseField  = "g-recaptcha-response"
	hCaptchaResponseField    = "h-captcha-response"
	mCaptchaResponseField    = "mcaptcha__token" // this form key is hard-coded in the mcaptcha frontend library
	cfTurnstileResponseField = "cf-turnstile-response"
)

// VerifyCaptcha returns whether the captcha is solved or disabled, otherwise it renders tpl with an error
func VerifyCaptcha(ctx *Context, tpl templates.TplName, form any) bool {
	if !setting.Service.EnableCaptcha {
		return true
	}

	var valid bool
	var err error
	switch setting.Service.CaptchaType {
	case setting.ImageCaptcha:
		valid = imagecaptcha.Verify(ctx.FormString(imageCaptchaIDField), ctx.FormString(imageCaptchaAnswerField))
	case setting.ReCaptcha:
		valid, err = recaptcha.Verify(ctx, ctx.Req.Form.Get(gRecaptchaResponseField))
	case setting.HCaptcha:
		valid, err = hcaptcha.Verify(ctx, ctx.Req.Form.Get(hCaptchaResponseField))
	case setting.MCaptcha:
		valid, err = mcaptcha.Verify(ctx, ctx.Req.Form.Get(mCaptchaResponseField))
	case setting.CfTurnstile:
		valid, err = turnstile.Verify(ctx, ctx.Req.Form.Get(cfTurnstileResponseField))
	default:
		ctx.ServerError("Unknown Captcha Type", fmt.Errorf("unknown Captcha Type: %s", setting.Service.CaptchaType))
		return false
	}
	if err != nil {
		log.Debug("Captcha Verify failed: %v", err)
	}

	if !valid {
		ctx.Data["Err_Captcha"] = true
		ctx.RenderWithErrDeprecated(ctx.Tr("form.captcha_incorrect"), tpl, form)
		return false
	}
	return true
}
