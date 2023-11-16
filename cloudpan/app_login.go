// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 电脑手机客户端API，例如MAC客户端
package cloudpan

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-api/cloudpan/apiutil"
	"github.com/tickstep/library-go/crypto"
	"github.com/tickstep/library-go/logger"
	"github.com/tickstep/library-go/requester"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type (
	appLoginParams struct {
		CaptchaToken string
		Lt           string
		ReturnUrl    string
		ParamId      string
		ReqId        string
		jRsaKey      string
		Pre          string
	}

	EncryptConfData struct {
		Pre    string `json:"pre"`
		PubKey string `json:"pubKey"`
	}
	EncryptConf struct {
		Result int             `json:"result"`
		Data   EncryptConfData `json:"data"`
	}

	AppLoginToken struct {
		SessionKey          string `json:"sessionKey"`
		SessionSecret       string `json:"sessionSecret"`
		FamilySessionKey    string `json:"familySessionKey"`
		FamilySessionSecret string `json:"familySessionSecret"`
		AccessToken         string `json:"accessToken"`
		RefreshToken        string `json:"refreshToken"`
		// 有效期的token
		SskAccessToken string `json:"sskAccessToken"`
		// token 过期时间点，时间戳ms
		SskAccessTokenExpiresIn int64  `json:"sskAccessTokenExpiresIn"`
		RsaPublicKey            string `json:"rsaPublicKey"`
	}

	appSessionResp struct {
		ResCode             int    `json:"res_code"`
		ResMessage          string `json:"res_message"`
		AccessToken         string `json:"accessToken"`
		FamilySessionKey    string `json:"familySessionKey"`
		FamilySessionSecret string `json:"familySessionSecret"`
		GetFileDiffSpan     int    `json:"getFileDiffSpan"`
		GetUserInfoSpan     int    `json:"getUserInfoSpan"`
		IsSaveName          string `json:"isSaveName"`
		KeepAlive           int    `json:"keepAlive"`
		LoginName           string `json:"loginName"`
		RefreshToken        string `json:"refreshToken"`
		SessionKey          string `json:"sessionKey"`
		SessionSecret       string `json:"sessionSecret"`
	}

	accessTokenResp struct {
		// token过期时间，默认30天
		ExpiresIn   int64  `json:"expiresIn"`
		AccessToken string `json:"accessToken"`
	}

	appRefreshUserSessionResp struct {
		XMLName             xml.Name `xml:"userSession"`
		LoginName           string   `xml:"loginName"`
		SessionKey          string   `xml:"sessionKey"`
		SessionSecret       string   `xml:"sessionSecret"`
		KeepAlive           int      `xml:"keepAlive"`
		GetFileDiffSpan     int      `xml:"getFileDiffSpan"`
		GetUserInfoSpan     int      `xml:"getUserInfoSpan"`
		FamilySessionKey    string   `xml:"familySessionKey"`
		FamilySessionSecret string   `xml:"familySessionSecret"`
	}
)

var (
	appClient = requester.NewHTTPClient()
)

func AppLogin(username, password string) (result *AppLoginToken, error *apierror.ApiError) {
	result = &AppLoginToken{}

	appClient.ResetCookiejar()
	loginParams, err := appGetLoginParams()
	if err != nil {
		logger.Verboseln("get login params error")
		return nil, err
	}
	rsaKey := &strings.Builder{}
	fmt.Fprintf(rsaKey, "-----BEGIN PUBLIC KEY-----\n%s\n-----END PUBLIC KEY-----", loginParams.jRsaKey)
	result.RsaPublicKey = rsaKey.String()
	rsaUserName, _ := crypto.RsaEncrypt([]byte(rsaKey.String()), []byte(username))
	rsaPassword, _ := crypto.RsaEncrypt([]byte(rsaKey.String()), []byte(password))

	needcaptchaMap := map[string]string{
		"accountType": "02",
		"appKey":      "8025431004",
		"userName":    loginParams.Pre + apiutil.B64toHex(string(crypto.Base64Encode(rsaUserName))),
	}
	needcaptchaDate, err1 := appClient.Fetch("POST", "https://open.e.189.cn/api/logbox/oauth2/needcaptcha.do", needcaptchaMap, nil)

	if err1 != nil {
		logger.Verboseln("login needcaptcha occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	needcaptchaStr := string(needcaptchaDate)
	if !strings.EqualFold("0", needcaptchaStr) {
		logger.Verboseln("login need captcha, stop login")
		return nil, apierror.NewApiErrorWithError(errors.New("login need captcha"))
	}

	urlStr := "https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do"
	headers := map[string]string{
		"Content-Type":     "application/x-www-form-urlencoded",
		"Referer":          "https://open.e.189.cn/api/logbox/oauth2/unifyAccountLogin.do",
		"Cookie":           "LT=" + loginParams.Lt,
		"X-Requested-With": "XMLHttpRequest",
		"REQID":            loginParams.ReqId,
		"lt":               loginParams.Lt,
	}
	formData := map[string]string{
		"appKey":       "8025431004",
		"accountType":  "02",
		"userName":     loginParams.Pre + apiutil.B64toHex(string(crypto.Base64Encode(rsaUserName))),
		"epd":          loginParams.Pre + apiutil.B64toHex(string(crypto.Base64Encode(rsaPassword))),
		"validateCode": "",
		"captchaToken": loginParams.CaptchaToken,
		"returnUrl":    loginParams.ReturnUrl,
		"mailSuffix":   "@189.cn",
		"dynamicCheck": "FALSE",
		"clientType":   "10020",
		"cb_SaveName":  "0",
		"isOauth2":     "false",
		"state":        "",
		"paramId":      loginParams.ParamId,
	}

	logger.Verboseln("do request url: " + urlStr)
	body, err1 := appClient.Fetch("POST", urlStr, formData, headers)
	if err1 != nil {
		logger.Verboseln("login redirectURL occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))
	r := &loginResult{}
	if err := json.Unmarshal(body, r); err != nil {
		logger.Verboseln("parse login result json error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}
	if r.Result != 0 || r.ToUrl == "" {
		return nil, apierror.NewFailedApiError("登录失败")
	}

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/getSessionForPC.action?clientType=%s&version=%s&channelId=%s&redirectURL=%s",
		API_URL, "TELEMAC", "1.0.0", "web_cloud.189.cn", url.QueryEscape(r.ToUrl))
	headers = map[string]string{
		"Accept": "application/json;charset=UTF-8",
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err1 = appClient.Fetch("GET", fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("get session info occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))
	rs := &appSessionResp{}
	if err := json.Unmarshal(body, rs); err != nil {
		logger.Verboseln("parse session result json error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}
	if rs.ResCode != 0 {
		return nil, apierror.NewFailedApiError("获取session失败")
	}
	result.SessionKey = rs.SessionKey
	result.SessionSecret = rs.SessionSecret
	result.FamilySessionKey = rs.FamilySessionKey
	result.FamilySessionSecret = rs.FamilySessionSecret
	result.AccessToken = rs.AccessToken
	result.RefreshToken = rs.RefreshToken

	// Ssk token
	fullUrl = &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/open/oauth2/getAccessTokenBySsKey.action?sessionKey=%s",
		API_URL, rs.SessionKey)
	timestamp := apiutil.Timestamp()
	signParams := map[string]string{
		"Timestamp":  strconv.Itoa(timestamp),
		"sessionKey": rs.SessionKey,
		"AppKey":     "601102120",
	}
	headers = map[string]string{
		"AppKey":    "601102120",
		"Signature": apiutil.SignatureOfMd5(signParams),
		"Sign-Type": "1",
		"Accept":    "application/json",
		"Timestamp": strconv.Itoa(timestamp),
	}
	body, err1 = appClient.Fetch("GET", fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("get accessToken occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))
	atr := &accessTokenResp{}
	if err := json.Unmarshal(body, atr); err != nil {
		logger.Verboseln("parse accessToken result json error ", err)
		return nil, apierror.NewFailedApiError(err.Error())
	}
	result.SskAccessTokenExpiresIn = atr.ExpiresIn
	result.SskAccessToken = atr.AccessToken
	return result, nil
}

func appGetLoginParams() (params appLoginParams, error *apierror.ApiError) {
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	fullUrl := &strings.Builder{}
	// use MAC client appid
	fmt.Fprintf(fullUrl, "%s/api/portal/unifyLoginForPC.action?appId=%s&clientType=%s&returnURL=%s&timeStamp=%d",
		WEB_URL, "8025431004", "10020", "https://m.cloud.189.cn/zhuanti/2020/loginErrorPc/index.html", apiutil.Timestamp())
	logger.Verboseln("do request url: " + fullUrl.String())
	data, err := appClient.Fetch("GET", fullUrl.String(), nil, header)
	if err != nil {
		logger.Verboseln("login redirectURL occurs error: ", err.Error())
		return params, apierror.NewApiErrorWithError(err)
	}
	content := string(data)

	re, _ := regexp.Compile("captchaToken' value='(.+?)'")
	params.CaptchaToken = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("lt = \"(.+?)\"")
	params.Lt = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("returnUrl = '(.+?)'")
	params.ReturnUrl = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("paramId = \"(.+?)\"")
	params.ParamId = re.FindStringSubmatch(content)[1]

	re, _ = regexp.Compile("reqId = \"(.+?)\"")
	params.ReqId = re.FindStringSubmatch(content)[1]

	//re, _ = regexp.Compile("j_rsaKey\" value=\"(.+?)\"")
	//params.jRsaKey = re.FindStringSubmatch(content)[1]

	formData := map[string]string{
		"appId": "8025431004",
	}
	// get rsa key
	data, err = appClient.Fetch("POST", "https://open.e.189.cn/api/logbox/config/encryptConf.do", formData, header)
	if err != nil {
		logger.Verboseln("get encryptConf occurs error: ", err.Error())
		return params, apierror.NewApiErrorWithError(err)
	}

	encryptConf := EncryptConf{}
	err = json.Unmarshal(data, &encryptConf)
	if err != nil {
		logger.Verboseln("Unmarshal json encryptConf occurs error: ", err.Error())
		return params, apierror.NewApiErrorWithError(err)
	}
	params.jRsaKey = encryptConf.Data.PubKey
	params.Pre = encryptConf.Data.Pre

	return
}

// getSessionByAccessToken 通过appSessionResp.accessToken刷新session信息
func getSessionByAccessToken(accessToken string) (*appRefreshUserSessionResp, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/getSessionForPC.action?appId=%s&accessToken=%s&clientSn=%s&%s",
		API_URL, "8025431004", accessToken, apiutil.Uuid(), apiutil.PcClientInfoSuffixParam())
	headers := map[string]string{
		"X-Request-ID": apiutil.XRequestId(),
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err1 := appClient.Fetch("GET", fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("getSessionByAccessToken occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))
	item := &appRefreshUserSessionResp{}
	if err := xml.Unmarshal(body, item); err != nil {
		logger.Verboseln("getSessionByAccessToken parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}
