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
	"strings"
)

type (
	// AppGetFileInfoParam 获取文件信息参数
	AppFamilyInfo struct {
		Count int `xml:"count" xml:"count"`
		Type int `xml:"type" json:"type"`
		UserRole int `xml:"userRole" json:"userRole"`
		CreateTime string `xml:"createTime" json:"createTime"`
		FamilyId int64 `xml:"familyId" json:"familyId"`
		RemarkName string `xml:"remarkName" json:"remarkName"`
		UseFlag int `xml:"useFlag" json:"useFlag"`
	}

	AppFamilyInfoListResult struct {
		XMLName xml.Name `xml:"familyListResponse"`
		FamilyInfoList []*AppFamilyInfo `xml:"familyInfo" json:"familyInfoList"`
	}

)

// AppGetFamilyList 获取用户的家庭列表
func (p *PanClient) AppFamilyGetFamilyList() (*AppFamilyInfoListResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/family/manage/getFamilyList.action?%s",
		API_URL, apiutil.PcClientInfoSuffixParam())
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	appToken := p.appToken
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.FamilySessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.FamilySessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppGetFamilyList occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "FamilyOperationFailed" {
			return nil, apierror.NewFailedApiError("获取家庭列表错误")
		}
	}
	item := &AppFamilyInfoListResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppGetFamilyList parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}