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
	"net/url"
	"strings"
)

type (
	AppMkdirResult struct {
		//XMLName xml.Name `xml:"folder"`
		// fileId 文件ID
		FileId string `xml:"id"`
		// ParentId 父文件夹ID
		ParentId string `xml:"parentId"`
		// FileName 名称
		FileName string `xml:"name"`
		// LastOpTime 最后修改时间
		LastOpTime string `xml:"lastOpTime"`
		// CreateTime 创建时间
		CreateTime string `xml:"createDate"`
		Rev        string `xml:"rev"`
		FileCata   int    `xml:"fileCata"`
	}
)

// AppMkdir 创建文件夹
func (p *PanClient) AppMkdir(familyId int64, parentFileId, dirName string) (*AppMkdirResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}

	sessionKey := ""
	sessionSecret := ""
	if familyId <= 0 {
		// 个人云
		fmt.Fprintf(fullUrl, "%s/createFolder.action?parentFolderId=%s&folderName=%s&relativePath=&%s",
			API_URL, parentFileId, url.QueryEscape(dirName), apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.SessionKey
		sessionSecret = p.appToken.SessionSecret
	} else {
		// 家庭云
		fmt.Fprintf(fullUrl, "%s/family/file/createFolder.action?familyId=%d&parentId=%s&folderName=%s&relativePath=&%s",
			API_URL, familyId, parentFileId, url.QueryEscape(dirName), apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.FamilySessionKey
		sessionSecret = p.appToken.FamilySessionSecret
	}
	httpMethod := "POST"
	dateOfGmt := apiutil.DateOfGmtStr()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppMkdir occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	item := &AppMkdirResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppMkdir parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}


func (p *PanClient) AppMkdirRecursive(familyId int64, parentFileId string, fullPath string, index int, pathSlice []string) (*AppMkdirResult, *apierror.ApiError) {
	r := &AppMkdirResult{}
	if familyId == 0 {
		if parentFileId == "" {
			// default root "/" entity
			parentFileId = NewAppFileEntityForRootDir().FileId
			if index == 0 && len(pathSlice) == 1 {
				// root path "/"
				r.FileId = parentFileId
				return r, nil
			}

			fullPath = ""
			return p.AppMkdirRecursive(familyId, parentFileId, fullPath, index + 1, pathSlice)
		}
	}

	if index >= len(pathSlice) {
		r.FileId = parentFileId
		return r, nil
	}

	listFilePath := NewAppFileListParam()
	listFilePath.FileId = parentFileId
	listFilePath.FamilyId = familyId
	fileResult, err := p.AppGetAllFileList(listFilePath)
	if err != nil {
		r.FileId = ""
		return r, err
	}

	// existed?
	for _, fileEntity := range fileResult.FileList {
		if fileEntity.FileName == pathSlice[index] {
			return p.AppMkdirRecursive(familyId, fileEntity.FileId, fullPath + "/" + pathSlice[index], index + 1, pathSlice)
		}
	}

	// not existed, mkdir dir
	name := pathSlice[index]
	if !apiutil.CheckFileNameValid(name) {
		r.FileId = ""
		return r, apierror.NewFailedApiError("文件夹名不能包含特殊字符：" + apiutil.FileNameSpecialChars)
	}

	if familyId > 0 {
		if parentFileId == "-11" {
			parentFileId = ""
		}
	}
	rs, err := p.AppMkdir(familyId, parentFileId, name)
	if err != nil {
		r.FileId = ""
		return r, err
	}

	if (index+1) >= len(pathSlice) {
		return rs, nil
	} else {
		return p.AppMkdirRecursive(familyId, rs.FileId, fullPath + "/" + pathSlice[index], index + 1, pathSlice)
	}
}

