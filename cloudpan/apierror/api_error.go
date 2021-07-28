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

package apierror

import "encoding/xml"

const (
	// 成功
	ApiCodeOk ApiCode = 0
	// 失败
	ApiCodeFailed ApiCode = 999

	// 验证码
	ApiCodeNeedCaptchaCode ApiCode = 10
	// 会话/Token已过期
	ApiCodeTokenExpiredCode ApiCode = 11
	// 文件不存在
	ApiCodeFileNotFoundCode ApiCode = 12
	// 上传文件失败
	ApiCodeUploadFileStatusVerifyFailed = 13
	// 上传文件数据偏移值校验失败
	ApiCodeUploadOffsetVerifyFailed = 14
	// 服务器上传文件不存在
	ApiCodeUploadFileNotFound = 15
	// 文件已存在
	ApiCodeFileAlreadyExisted = 16
	// 上传达到日数量上限
	ApiCodeUserDayFlowOverLimited = 17
	// 参数无效，或者token过期
	ApiCodeInvalidArgument = 18
	// 敏感文件，禁止上传
	ApiCodeInfoSecurityError = 19
)

type ApiCode int

type ApiError struct {
	Code ApiCode
	Err string
}

func NewApiError(code ApiCode, err string) *ApiError {
	return &ApiError {
		code,
		err,
	}
}

func NewApiErrorWithError(err error) *ApiError {
	if err == nil {
		return NewApiError(ApiCodeOk, "")
	} else {
		return NewApiError(ApiCodeFailed, err.Error())
	}
}

func NewOkApiError() *ApiError {
	return NewApiError(ApiCodeOk, "")
}

func NewFailedApiError(err string) *ApiError {
	return NewApiError(ApiCodeFailed, err)
}

func (a *ApiError) SetErr(code ApiCode, err string) {
	a.Code = code
	a.Err = err
}

func (a *ApiError) Error() string {
	return a.Err
}

func (a *ApiError) ErrCode() ApiCode {
	return a.Code
}

// ParseAppCommonApiError 解析公共错误，如果没有错误则返回nil
func ParseAppCommonApiError(data []byte) *ApiError  {
	errResp := &AppErrorXmlResp{}
	if err := xml.Unmarshal(data, errResp); err == nil {
		if errResp.Code != "" {
			if "InvalidArgument" == errResp.Code {
				return NewApiError(ApiCodeInvalidArgument, "参数无效")
			} else if "InfoSecurityErrorCode" == errResp.Code {
				return NewApiError(ApiCodeInfoSecurityError, "敏感文件或受版权保护，禁止上传")
			} else if "UserDayFlowOverLimited" == errResp.Code {
				return NewApiError(ApiCodeUserDayFlowOverLimited, "账号上传达到每日数量限额")
			}
			return NewFailedApiError(errResp.Message)
		}
	}
	return nil
}