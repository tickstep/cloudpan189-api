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

package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-api/cloudpan/apiutil"
	"github.com/tickstep/library-go/logger"
	"net/url"
	"strconv"
)

type (
	AppInitMultiUploadParam struct {
		// FamilyId 家庭ID。如果是0代表是个人云
		FamilyId int64
		// ParentFolderId 存储云盘的目录ID
		ParentFolderId string
		// FileName 存储云盘的文件名
		FileName string
		// Size 文件总大小
		Size int64
		// Md5 文件MD5
		Md5 string
		// SliceSize 分片大小
		SliceSize int64
		// 第一分片的MD5
		SliceMd5 string
		// LastWrite 文件最后修改日期，格式：2018-11-18 09:12:13
		LastWrite string
		// LocalPath 文件存储的本地绝对路径
		LocalPath string
	}

	AppInitMultiUploadResult struct {
		UploadType   int    `json:"uploadType"`
		UploadHost   string `json:"uploadHost"`
		UploadFileId string `json:"uploadFileId"`
		// FileDataExists 0-不存在， 1-已存在，可以秒传
		FileDataExists int `json:"fileDataExists"`
	}
)

func (a *AppInitMultiUploadParam) isFamily() bool {
	return a.FamilyId > 0
}

// AppInitMultiUpload 创建预上传
func (p *PanClient) AppInitMultiUpload(param *AppInitMultiUploadParam) (*AppInitMultiUploadResult, *apierror.ApiError) {
	fullUrl := UPLOAD_URL
	if param.isFamily() {
		fullUrl += "/family"
	} else {
		fullUrl += "/person"
	}
	paramData := Params{
		"parentFolderId": param.ParentFolderId,
		"fileName":       url.QueryEscape(param.FileName),
		"fileSize":       fmt.Sprint(param.Size),
		"fileMd5":        param.Md5,
		"sliceSize":      fmt.Sprint(param.SliceSize),
		"sliceMd5":       param.SliceMd5,
	}
	if param.isFamily() {
		paramData.Set("familyId", strconv.FormatInt(param.FamilyId, 10))
	}

	// 查询参数
	paramStr := p.EncryptParams(paramData)
	if paramStr != "" {
		fullUrl += "/initMultiUpload?params=" + paramStr + "&" + apiutil.PcClientInfoSuffixParam()
	} else {
		fullUrl += "/initMultiUpload?" + apiutil.PcClientInfoSuffixParam()
	}

	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := apiutil.XRequestId()
	sessionKey := p.appToken.SessionKey
	sessionSecret := p.appToken.SessionSecret
	if param.isFamily() {
		sessionKey = p.appToken.FamilySessionKey
		sessionSecret = p.appToken.FamilySessionSecret
	}
	headers := map[string]string{
		"isjson":       "1",
		"Date":         dateOfGmt,
		"SessionKey":   sessionKey,
		"Signature":    apiutil.SignatureOfHmacV2(sessionSecret, sessionKey, httpMethod, fullUrl, dateOfGmt, paramStr),
		"X-Request-ID": requestId,
	}

	logger.Verboseln("do request url: " + fullUrl)
	body, err1 := p.client.Fetch(httpMethod, fullUrl, nil, headers)
	if err1 != nil {
		logger.Verboseln("AppInitMultiUpload occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))

	// handler common error
	if apiErr := apierror.ParseAppJsonCommonApiError(body); apiErr != nil {
		return nil, apiErr
	}

	var r struct {
		Code string                    `json:"code"`
		Data *AppInitMultiUploadResult `json:"data"`
	}
	if err := json.Unmarshal(body, &r); err != nil {
		logger.Verboseln("AppInitMultiUpload parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return r.Data, nil
}
