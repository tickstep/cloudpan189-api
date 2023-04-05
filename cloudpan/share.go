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
	"github.com/tickstep/library-go/text"
	"net/url"
	"strconv"
	"strings"
)

type (
	ShareExpiredTime int
	ShareMode        int

	PrivateShareResult struct {
		AccessCode    string `json:"accessCode"`
		ShortShareUrl string `json:"shortShareUrl"`
	}

	PublicShareResult struct {
		ShareId       int64  `json:"shareId"`
		ShortShareUrl string `json:"shortShareUrl"`
	}

	AccessCount struct {
		CopyCount     int `json:"copyCount"`
		DownloadCount int `json:"downloadCount"`
		PreviewCount  int `json:"previewCount"`
	}

	ShareItem struct {
		// AccessCode 提取码，私密分享才有
		AccessCode string `json:"accessCode"`
		// AccessURL 分享链接
		AccessURL string `json:"accessURL"`
		// AccessCount 分享被查看下载次数
		AccessCount AccessCount `json:"accessCount"`
		// DownloadUrl 下载路径，文件才会有
		DownloadUrl string `json:"downloadUrl"`
		// DownloadUrl 下载路径，文件才会有
		LongDownloadUrl string `json:"longDownloadUrl"`
		// FileId 文件ID
		FileId string `json:"fileId"`
		// FileIdDigest 文件指纹
		FileIdDigest string `json:"fileIdDigest"`
		// FileName 文件名
		FileName string `json:"fileName"`
		// FilePath 路径
		FilePath string `json:"filePath"`
		// FileSize 文件大小，文件夹为0
		FileSize int64 `json:"fileSize"`
		// IconURL 缩略图路径???
		IconURL string `json:"iconURL"`
		// IsFolder 是否是文件夹
		IsFolder bool `json:"isFolder"`
		// MediaType 文件类别
		MediaType      MediaType `json:"mediaType"`
		NeedAccessCode int       `json:"needAccessCode"`
		// NickName 分享者账号昵称
		NickName string `json:"nickName"`
		// ReviewStatus 审查状态，1-正常
		ReviewStatus int `json:"reviewStatus"`
		// ShareDate 分享日期
		ShareDate int64 `json:"shareDate"`
		// ShareId 分享项目ID，唯一标识该分享项
		ShareId int64 `json:"shareId"`
		// ShareMode 分享模式，1-私密，2-公开
		ShareMode ShareMode `json:"shareMode"`
		// ShareTime 分享时间
		ShareTime int64 `json:"shareTime"`
		// ShareType 分享类别，默认都是1
		ShareType int `json:"shareType"`
		// ShortShareUrl 分享的访问路径，和 AccessURL 一致
		ShortShareUrl string `json:"shortShareUrl"`
	}

	ShareItemList []*ShareItem

	// ShareListResult 获取分享项目列表响应体
	ShareListResult struct {
		Data        ShareItemList `json:"data"`
		PageNum     int           `json:"pageNum"`
		PageSize    int           `json:"pageSize"`
		RecordCount int           `json:"recordCount"`
	}

	ShareListParam struct {
		ShareType int `json:"shareType"`
		PageNum   int `json:"pageNum"`
		PageSize  int `json:"pageSize"`
	}

	errResp struct {
		ErrorVO apierror.ErrorResp `json:"errorVO"`
	}

	// 转存分享
	listShareDirResult struct {
		ResCode    int    `json:"res_code"`
		ResMessage string `json:"res_message"`
		ExpireTime int    `json:"expireTime"`
		ExpireType int    `json:"expireType"`
		FileListAO struct {
			Count    int `json:"count"`
			FileList []struct {
				CreateDate string `json:"createDate"`
				FileCata   int    `json:"fileCata"`
				Id         int64  `json:"id"`
				LastOpTime string `json:"lastOpTime"`
				Md5        string `json:"md5"`
				MediaType  int    `json:"mediaType"`
				Name       string `json:"name"`
				Rev        string `json:"rev"`
				Size       int64  `json:"size"`
				StarLabel  int    `json:"starLabel"`
			} `json:"fileList"`
			FileListSize int64 `json:"fileListSize"`
			FolderList   []struct {
				CreateDate   string `json:"createDate"`
				FileCata     int    `json:"fileCata"`
				FileListSize int    `json:"fileListSize"`
				Id           int64  `json:"id"`
				LastOpTime   string `json:"lastOpTime"`
				Name         string `json:"name"`
				ParentId     int64  `json:"parentId"`
				Rev          string `json:"rev"`
				StarLabel    int    `json:"starLabel"`
			} `json:"folderList"`
		} `json:"fileListAO"`
		LastRev int64 `json:"lastRev"`
	}
)

const (
	// 1天期限
	ShareExpiredTime1Day ShareExpiredTime = 1
	// 7天期限
	ShareExpiredTime7Day ShareExpiredTime = 7
	// 永久期限
	ShareExpiredTimeForever ShareExpiredTime = 2099

	// ShareModePrivate 私密分享
	ShareModePrivate ShareMode = 1
	// ShareModePublic 公开分享
	ShareModePublic ShareMode = 2
)

func (p *PanClient) SharePrivate(fileId string, expiredTime ShareExpiredTime) (*PrivateShareResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/api/open/share/createShareLink.action?fileId=%s&expireTime=%d&shareType=3",
		WEB_URL, fileId, expiredTime)
	logger.Verboseln("do request url: " + fullUrl.String())
	//body, err := p.client.DoGet(fullUrl.String())
	headers := map[string]string{
		"accept": "application/json;charset=UTF-8",
	}
	body, err := p.client.Fetch("GET", fullUrl.String(), nil, headers)
	logger.Verboseln("response body: " + string(body))
	if err != nil {
		logger.Verboseln("SharePrivate failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	errResp := &errResp{}
	if err := json.Unmarshal(body, errResp); err == nil {
		if errResp.ErrorVO.ErrorCode != "" {
			logger.Verboseln("SharePrivate response failed")
			if errResp.ErrorVO.ErrorCode == "ShareCreateOverload" {
				return nil, apierror.NewFailedApiError("您分享的次数已达上限，请明天再来吧")
			}
			return nil, apierror.NewApiErrorWithError(err)
		}
	}

	type shareLink struct {
		AccessCode string `json:"accessCode"`
		AccessUrl  string `json:"accessUrl"`
		FileId     int64  `json:"fileId"`
		ShareId    int64  `json:"shareId"`
		Url        string `json:"url"`
	}
	type shareLinkResult struct {
		Code          int         `json:"res_code"`
		Message       string      `json:"res_message"`
		ShareLinkList []shareLink `json:"shareLinkList"`
	}
	r := shareLinkResult{}
	if err := json.Unmarshal(body, &r); err != nil {
		logger.Verboseln("SharePrivate response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return &PrivateShareResult{
		AccessCode:    r.ShareLinkList[0].AccessCode,
		ShortShareUrl: r.ShareLinkList[0].Url,
	}, nil
}

func (p *PanClient) SharePublic(fileId string, expiredTime ShareExpiredTime) (*PublicShareResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/createOutLinkShare.action?fileId=%s&expireTime=%d&withAccessCode=1",
		WEB_URL, fileId, expiredTime)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("SharePublic failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &PublicShareResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("SharePublic response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

func NewShareListParam() *ShareListParam {
	return &ShareListParam{
		ShareType: 1,
		PageNum:   1,
		PageSize:  60,
	}
}
func (p *PanClient) ShareList(param *ShareListParam) (*ShareListResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/api/portal/listShares.action?shareType=%d&pageNum=%d&pageSize=%d",
		WEB_URL, param.ShareType, param.PageNum, param.PageSize)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("ShareList failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &ShareListResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("ShareList response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	// normalize
	for _, s := range item.Data {
		s.AccessURL = "https:" + s.AccessURL
		s.ShortShareUrl = "https:" + s.ShortShareUrl
	}
	return item, nil
}

func (p *PanClient) ShareCancel(shareIdList []int64) (bool, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	shareIds := ""
	for _, id := range shareIdList {
		shareIds += strconv.FormatInt(id, 10) + ","
	}
	if strings.LastIndex(shareIds, ",") == (len(shareIds) - 1) {
		shareIds = text.Substr(shareIds, 0, len(shareIds)-1)
	}

	fmt.Fprintf(fullUrl, "%s/api/portal/cancelShare.action?shareIdList=%s&ancelType=1",
		WEB_URL, url.QueryEscape(shareIds))
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("ShareCancel failed")
		return false, apierror.NewApiErrorWithError(err)
	}
	comResp := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, comResp); err == nil {
		if comResp.ErrorCode != "" {
			logger.Verboseln("ShareCancel response failed")
			return false, apierror.NewFailedApiError("取消分享失败，请稍后重试")
		}
	}
	item := &apierror.SuccessResp{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("ShareCancel response failed")
		return false, apierror.NewApiErrorWithError(err)
	}
	return item.Success, nil
}

// ShareSave 转存分享到对应的文件夹
func (p *PanClient) ShareSave(accessUrl string, accessCode string, savePanDirId string) (bool, *apierror.ApiError) {
	shareCode := ""
	idx := strings.LastIndex(accessUrl, "/")
	if idx > 0 {
		rs := []rune(accessUrl)
		shareCode = string(rs[idx+1:])
	}
	fullUrl := &strings.Builder{}
	header := map[string]string{
		"accept":     "application/json;charset=UTF-8",
		"origin":     "https://cloud.189.cn",
		"Referer":    "https://cloud.189.cn/web/share?code=" + shareCode,
		"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_3_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.96 Safari/537.36",
	}

	// 获取分享基础信息
	fmt.Fprintf(fullUrl, "%s/api/open/share/getShareInfoByCode.action?&shareCode=%s",
		WEB_URL, shareCode)

	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := client.Fetch("GET", fullUrl.String(), nil, header)
	if err != nil {
		logger.Verboseln("ShareListDirDetail failed")
		return false, apierror.NewApiErrorWithError(err)
	}

	type shareInfoByCode struct {
		ResCode        int    `json:"res_code"`
		ResMessage     string `json:"res_message"`
		AccessCode     string `json:"accessCode"`
		ExpireTime     int    `json:"expireTime"`
		ExpireType     int    `json:"expireType"`
		FileId         string `json:"fileId"`
		FileName       string `json:"fileName"`
		FileSize       int    `json:"fileSize"`
		IsFolder       bool   `json:"isFolder"`
		NeedAccessCode int    `json:"needAccessCode"`
		ShareDate      int64  `json:"shareDate"`
		ShareId        int64  `json:"shareId"`
		ShareMode      int    `json:"shareMode"`
		ShareType      int    `json:"shareType"`
	}
	shareInfoEnity := &shareInfoByCode{}
	if err := json.Unmarshal(body, shareInfoEnity); err != nil {
		logger.Verboseln("getShareInfoByCode response failed")
		return false, apierror.NewApiErrorWithError(err)
	}

	// 获取分享文件列表
	fullUrl = &strings.Builder{}
	if shareInfoEnity.IsFolder {
		fmt.Fprintf(fullUrl, "%s/api/open/share/listShareDir.action?pageNum=1&pageSize=60&fileId=%s&shareDirFileId=%s&isFolder=true&shareId=%d&shareMode=%d&iconOption=5&orderBy=lastOpTime&descending=true&accessCode=%s",
			WEB_URL, shareInfoEnity.FileId, shareInfoEnity.FileId, shareInfoEnity.ShareId, shareInfoEnity.ShareMode, accessCode)
	} else {
		fmt.Fprintf(fullUrl, "%s/api/open/share/listShareDir.action?fileId=%s&shareId=%d&shareMode=%d&isFolder=false&iconOption=5&pageNum=1&pageSize=10&accessCode=%s",
			WEB_URL, shareInfoEnity.FileId, shareInfoEnity.ShareId, shareInfoEnity.ShareMode, accessCode)
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err = client.Fetch("GET", fullUrl.String(), nil, header)
	if err != nil {
		logger.Verboseln("listShareDir failed")
		return false, apierror.NewApiErrorWithError(err)
	}

	listShareDirEnity := &listShareDirResult{}
	if err := json.Unmarshal(body, listShareDirEnity); err != nil {
		logger.Verboseln("listShareDir response failed")
		return false, apierror.NewApiErrorWithError(err)
	}

	// 转存分享
	taskReqParam := &BatchTaskParam{
		TypeFlag:       BatchTaskTypeShareSave,
		TaskInfos:      makeBatchTaskInfoListForShareSave(listShareDirEnity),
		TargetFolderId: savePanDirId,
		ShareId:        shareInfoEnity.ShareId,
	}
	taskId, apierror1 := p.CreateBatchTask(taskReqParam)
	logger.Verboseln("share save taskid: ", taskId)
	return taskId != "", apierror1
}

func makeBatchTaskInfoListForShareSave(opFileList *listShareDirResult) (infoList BatchTaskInfoList) {
	// file
	for _, fe := range opFileList.FileListAO.FileList {
		infoItem := &BatchTaskInfo{
			FileId:   strconv.FormatInt(fe.Id, 10),
			FileName: fe.Name,
			IsFolder: 0,
		}
		infoList = append(infoList, infoItem)
	}

	// folder
	for _, fe := range opFileList.FileListAO.FolderList {
		infoItem := &BatchTaskInfo{
			FileId:   strconv.FormatInt(fe.Id, 10),
			FileName: fe.Name,
			IsFolder: 1,
		}
		infoList = append(infoList, infoItem)
	}
	return
}
