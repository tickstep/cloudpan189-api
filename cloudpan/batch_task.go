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
	"github.com/tickstep/library-go/logger"
	"strconv"
	"strings"
)

type (
	// TaskInfo 任务信息
	BatchTaskInfo struct {
		// FileId 文件ID
		FileId string `json:"fileId"`
		// FileName 文件名
		FileName string `json:"fileName"`
		// IsFolder 是否是文件夹，0-否，1-是
		IsFolder int `json:"isFolder"`
		// SrcParentId 文件所在父目录ID
		SrcParentId string `json:"srcParentId"`
	}

	BatchTaskInfoList []*BatchTaskInfo

	// BatchTaskParam 任务参数
	BatchTaskParam struct {
		TypeFlag       BatchTaskType     `json:"type"`
		TaskInfos      BatchTaskInfoList `json:"taskInfos"`
		TargetFolderId string            `json:"targetFolderId"`
		ShareId        int64             `json:"shareId"`
	}

	// CheckTaskResult 检查任务结果
	CheckTaskResult struct {
		FailedCount         int     `json:"failedCount" xml:"failedCount"`
		SkipCount           int     `json:"skipCount" xml:"skipCount"`
		SubTaskCount        int     `json:"subTaskCount" xml:"subTaskCount"`
		SuccessedCount      int     `json:"successedCount" xml:"successedCount"`
		SuccessedFileIdList []int64 `json:"successedFileIdList" xml:"successedFileIdList"`
		TaskId              string  `json:"taskId" xml:"taskId"`
		// TaskStatus 任务状态， 4-成功
		TaskStatus BatchTaskStatus `json:"taskStatus" xml:"taskStatus"`
	}

	BatchTaskStatus int
	BatchTaskType   string
)

const (
	// BatchTaskStatusNotAction 无需任何操作
	BatchTaskStatusNotAction BatchTaskStatus = 2
	// BatchTaskStatusOk 成功
	BatchTaskStatusOk BatchTaskStatus = 4

	// BatchTaskTypeDelete 删除文件任务
	BatchTaskTypeDelete BatchTaskType = "DELETE"
	// BatchTaskTypeCopy 复制文件任务
	BatchTaskTypeCopy BatchTaskType = "COPY"
	// BatchTaskTypeMove 移动文件任务
	BatchTaskTypeMove BatchTaskType = "MOVE"

	// BatchTaskTypeRecycleRestore 还原回收站文件
	BatchTaskTypeRecycleRestore BatchTaskType = "RESTORE"

	// BatchTaskTypeShareSave 转录分享
	BatchTaskTypeShareSave BatchTaskType = "SHARE_SAVE"
)

func (p *PanClient) CreateBatchTask(param *BatchTaskParam) (taskId string, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	//fmt.Fprintf(fullUrl, "%s/createBatchTask.action", WEB_URL)
	fmt.Fprintf(fullUrl, "%s/api/open/batch/createBatchTask.action", WEB_URL)
	logger.Verboseln("do request url: " + fullUrl.String())
	taskInfosStr, err := json.Marshal(param.TaskInfos)
	var postData map[string]string
	if BatchTaskTypeDelete == param.TypeFlag || BatchTaskTypeRecycleRestore == param.TypeFlag {
		postData = map[string]string{
			"type":      string(param.TypeFlag),
			"taskInfos": string(taskInfosStr),
		}
	} else if BatchTaskTypeCopy == param.TypeFlag || BatchTaskTypeMove == param.TypeFlag {
		postData = map[string]string{
			"type":           string(param.TypeFlag),
			"taskInfos":      string(taskInfosStr),
			"targetFolderId": param.TargetFolderId,
		}
	} else if BatchTaskTypeShareSave == param.TypeFlag {
		type batchTaskShareSaveInfo struct {
			// FileId 文件ID
			FileId string `json:"fileId"`
			// FileName 文件名
			FileName string `json:"fileName"`
			// IsFolder 是否是文件夹，0-否，1-是
			IsFolder int `json:"isFolder"`
		}
		tsl := []*batchTaskShareSaveInfo{}
		for _, item := range param.TaskInfos {
			tsl = append(tsl, &batchTaskShareSaveInfo{
				FileId:   item.FileId,
				FileName: item.FileName,
				IsFolder: item.IsFolder,
			})
		}
		taskInfosStr, _ = json.Marshal(tsl)
		postData = map[string]string{
			"type":           string(param.TypeFlag),
			"taskInfos":      string(taskInfosStr),
			"targetFolderId": param.TargetFolderId,
			"shareId":        strconv.FormatInt(param.ShareId, 10),
		}
	} else {
		return "", apierror.NewFailedApiError("不支持的操作")
	}

	//body, err := p.client.DoPost(fullUrl.String(), postData)
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"accept":       "application/json;charset=UTF-8",
	}
	body, err := p.client.Fetch("POST", fullUrl.String(), postData, headers)
	if err != nil {
		logger.Verboseln("CreateBatchTask failed")
		return "", apierror.NewApiErrorWithError(err)
	}
	comResp := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, comResp); err == nil {
		if comResp.ErrorCode == "InternalError" {
			logger.Verboseln("response failed", comResp)
			return "", apierror.NewFailedApiError("操作失败")
		}
	}
	type TaskResp struct {
		ErrorCode int64  `json:"res_code"`
		ErrorMsg  string `json:"res_message"`
		TaskId    string `json:"taskId"`
	}
	t := &TaskResp{}
	if err1 := json.Unmarshal([]byte(body), t); err1 == nil {
		return t.TaskId, nil
	}
	return "", nil
}

func (p *PanClient) CheckBatchTask(typeFlag BatchTaskType, taskId string) (result *CheckTaskResult, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/api/open/batch/checkBatchTask.action", WEB_URL)
	logger.Verboseln("do request url: " + fullUrl.String())
	postData := map[string]string{
		"type":   string(typeFlag),
		"taskId": taskId,
	}
	//body, err := p.client.DoPost(fullUrl.String(), postData)
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"accept":       "application/json;charset=UTF-8",
	}
	body, err := p.client.Fetch("POST", fullUrl.String(), postData, headers)
	if err != nil {
		logger.Verboseln("CheckBatchTask failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &CheckTaskResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("CheckBatchTask response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}
