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
	"encoding/xml"
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-api/cloudpan/apiutil"
	"github.com/tickstep/library-go/logger"
	"strconv"
	"strings"
)

func (p *PanClient) AppFamilyGetFileDownloadUrl(familyId int64, fileId string) (string, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	appToken := p.appToken
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	fmt.Fprintf(fullUrl, "%s/family/file/getFileDownloadUrl.action?familyId=%d&fileId=%s&%s",
		API_URL, familyId, fileId, apiutil.PcClientInfoSuffixParam())
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppGetFileDownloadUrl occurs error: ", err1.Error())
		return "", apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))

	type fdUrl struct {
		XMLName xml.Name `xml:"fileDownloadUrl"`
		FileDownloadUrl string `xml:",innerxml"`
	}

	item := &fdUrl{}
	if err := xml.Unmarshal(body, item); err != nil {
		fmt.Println("AppGetFileDownloadUrl parse response failed")
		return "", apierror.NewApiErrorWithError(err)
	}
	return strings.ReplaceAll(item.FileDownloadUrl, "&amp;", "&"), nil
}

func (p *PanClient) AppFamilyDownloadFileData(downloadFileUrl string, fileRange AppFileDownloadRange, downloadFunc DownloadFuncCallback) *apierror.ApiError {
	fullUrl := &strings.Builder{}

	fmt.Fprintf(fullUrl, "%s&%s",
		downloadFileUrl, apiutil.PcClientInfoSuffixParam())

	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := apiutil.XRequestId()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": requestId,
	}

	// 支持断点续传
	if fileRange.Offset != 0 || fileRange.End != 0 {
		rangeStr := "bytes=" + strconv.FormatInt(fileRange.Offset, 10) + "-"
		if fileRange.End != 0 {
			rangeStr += strconv.FormatInt(fileRange.End, 10)
		}
		headers["range"] = rangeStr
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	_, err := downloadFunc(httpMethod, fullUrl.String(), headers)
	//resp, err := p.client.Req(httpMethod, fullUrl.String(), nil, headers)
	if err != nil {
		logger.Verboseln("AppDownloadFileData response failed")
		return apierror.NewApiErrorWithError(err)
	}
	return nil
}