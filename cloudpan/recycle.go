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
	"net/url"
	"strings"
)

type (
	// RecycleFileInfo 回收站中文件/目录信息
	RecycleFileInfo struct {
		// CreateTime 创建时间
		CreateDate string `json:"createDate"`
		// FileId 文件ID
		FileId int64 `json:"id"`
		// FileName 文件名
		FileName string `json:"name"`
		// FileSize 文件大小，文件夹为0
		FileSize int64 `json:"size"`
		// LastOpTime 最后修改时间
		LastOpTime string `json:"lastOpTime"`
		// ParentId 父文件ID
		Md5 string `json:"md5"`
		// MediaType 媒体类型
		MediaType int `json:"mediaType"`
		// PathStr 文件的完整路径
		PathStr string `json:"pathStr"`
	}

	RecycleFileInfoList []*RecycleFileInfo

	RecycleFileListResult struct {
		// Data 数据
		FileList RecycleFileInfoList `json:"fileList"`
		// RecordCount 文件总数量
		Count   uint   `json:"count"`
		Code    int64  `json:"res_code"`
		Message string `json:"res_message"`
	}

	RecycleFileActResult struct {
		Success bool `json:"success"`
	}
)

// RecycleList 列出回收站文件列表
func (p *PanClient) RecycleList(pageNum, pageSize int) (result *RecycleFileListResult, error *apierror.ApiError) {
	if pageNum <= 1 {
		pageNum = 1
	}
	if pageSize <= 1 {
		pageSize = 60
	}
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/api/open/file/listRecycleBinFiles.action?pageNum=%d&pageSize=%d&iconOption=1&family=false",
		WEB_URL, pageNum, pageSize)
	logger.Verboseln("do request url: " + fullUrl.String())
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"accept":       "application/json;charset=UTF-8",
	}
	body, err := p.client.Fetch("GET", fullUrl.String(), nil, headers)
	if err != nil {
		logger.Verboseln("RecycleList failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &RecycleFileListResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("RecycleList response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

// RecycleDelete 删除回收站文件或目录
func (p *PanClient) RecycleDelete(familyId int64, fileIdList []string) *apierror.ApiError {
	fullUrl := &strings.Builder{}
	if fileIdList == nil {
		return nil
	}
	if familyId <= 0 {
		fmt.Fprintf(fullUrl, "%s/v2/deleteFile.action?fileIdList=%s",
			WEB_URL, url.QueryEscape(strings.Join(fileIdList, ",")))
	} else {
		fmt.Fprintf(fullUrl, "%s/v2/deleteFile.action?familyId=%d&fileIdList=%s",
			WEB_URL, familyId, url.QueryEscape(strings.Join(fileIdList, ",")))
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("RecycleDelete failed")
		return apierror.NewApiErrorWithError(err)
	}
	item := &RecycleFileActResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("RecycleDelete response failed")
		return apierror.NewApiErrorWithError(err)
	}
	if !item.Success {
		return apierror.NewFailedApiError("failed")
	}
	return nil
}

func (p *PanClient) RecycleRestore(fileList []*RecycleFileInfo) (taskId string, err *apierror.ApiError) {
	if fileList == nil {
		return "", nil
	}

	taskReqParam := &BatchTaskParam{
		TypeFlag:  BatchTaskTypeRecycleRestore,
		TaskInfos: makeBatchTaskInfoList(fileList),
	}
	return p.CreateBatchTask(taskReqParam)
}

func makeBatchTaskInfoList(opFileList []*RecycleFileInfo) (infoList BatchTaskInfoList) {
	//for _, fe := range opFileList {
	//	isFolder := 0
	//	if fe.IsFolder {
	//		isFolder = 1
	//	}
	//	infoItem := &BatchTaskInfo{
	//		FileId:      fe.FileId,
	//		FileName:    fe.FileName,
	//		IsFolder:    isFolder,
	//		SrcParentId: fe.ParentId,
	//	}
	//	infoList = append(infoList, infoItem)
	//}
	return
}

func (p *PanClient) RecycleClear(familyId int64) *apierror.ApiError {
	fullUrl := &strings.Builder{}
	if familyId <= 0 {
		fmt.Fprintf(fullUrl, "%s/v2/emptyRecycleBin.action",
			WEB_URL)
	} else {
		fmt.Fprintf(fullUrl, "%s/v2/emptyRecycleBin.action?familyId=%d",
			WEB_URL, familyId)
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("RecycleClear failed")
		return apierror.NewApiErrorWithError(err)
	}
	item := &RecycleFileActResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("RecycleClear response failed")
		return apierror.NewApiErrorWithError(err)
	}
	if !item.Success {
		return apierror.NewFailedApiError("failed")
	}
	return nil
}
