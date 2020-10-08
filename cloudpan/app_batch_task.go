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
	"encoding/xml"
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-api/cloudpan/apiutil"
	"github.com/tickstep/library-go/logger"
	"strconv"
	"strings"
)

type (
	AppCreateBatchTaskResult struct {
		TaskId string `xml:"taskId"`
	}
	
	AppCheckBatchTaskResult struct {
		TaskId string `xml:"taskId"`
		TaskStatus int `xml:"taskStatus"`
		SubTaskCount int `xml:"subTaskCount"`
		SuccessCount int `xml:"successedCount"`
		FailedCount int `xml:"failedCount"`
		SkipCount int `xml:"skipCount"`
	}
)

// AppCreateBatchTask 创建批量处理任务
func (p *PanClient) AppCreateBatchTask(familyId int64, param *BatchTaskParam) (taskId string, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}

	fmt.Fprintf(fullUrl, "%s/batch/createBatchTask.action", API_URL)
	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "POST"
	dateOfGmt := apiutil.DateOfGmtStr()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	taskInfosStr, err := json.Marshal(param.TaskInfos)
	var postData map[string]string
	if BatchTaskTypeDelete == param.TypeFlag {
		postData = map[string]string {
			"type": string(param.TypeFlag),
			"taskInfos": string(taskInfosStr),
		}
	} else {
		return "", apierror.NewFailedApiError("不支持的操作")
	}

	// add common parameters
	postData["familyId"] = strconv.FormatInt(familyId, 10)
	postData["clientType"] = "TELEPC"
	postData["version"] = "6.2"
	postData["channelId"] = "web_cloud.189.cn"
	postData["rand"] = apiutil.Rand()

	respBody, err := p.client.Fetch(httpMethod, fullUrl.String(), postData, headers)
	if err != nil {
		logger.Verboseln("AppCreateBatchTask failed")
		return "", apierror.NewApiErrorWithError(err)
	}
	logger.Verboseln("response: " + string(respBody))

	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "InternalError" {
				return "", apierror.NewFailedApiError("内部错误")
			}
			return "", apierror.NewFailedApiError("请求出错")
		}
	}

	item := &AppCreateBatchTaskResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppCreateBatchTask parse response failed")
		return "", apierror.NewApiErrorWithError(err)
	}
	return item.TaskId, nil
}

// AppCheckBatchTask 检测批量任务状态和结果
func (p *PanClient) AppCheckBatchTask (typeFlag BatchTaskType, taskId string) (result *CheckTaskResult, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/batch/checkBatchTask.action", API_URL)
	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "POST"
	dateOfGmt := apiutil.DateOfGmtStr()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	postData := map[string]string {
		"type": string(typeFlag),
		"taskId": taskId,
		"clientType": "TELEPC",
		"version": "6.2",
		"channelId": "web_cloud.189.cn",
		"rand": apiutil.Rand(),
	}
	respBody, err := p.client.Fetch(httpMethod, fullUrl.String(), postData, headers)
	if err != nil {
		logger.Verboseln("AppCheckBatchTask failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	logger.Verboseln("response: " + string(respBody))

	item := &CheckTaskResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppCheckBatchTask response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}