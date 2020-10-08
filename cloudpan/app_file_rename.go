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

// AppRenameFile 重命名文件/文件夹
func (p *PanClient) AppRenameFile(renameFileId, newName string) (*AppFileEntity, *apierror.ApiError) {
	return p.appRenameFileInternal(renameFileId, newName, false)
}

func (p *PanClient) appRenameFileInternal(renameFileId, newName string, isFolder bool) (*AppFileEntity, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	if isFolder {
		fmt.Fprintf(fullUrl, "%s/renameFile.action?folderId=%s&destFolderName=%s&%s",
			API_URL,
			renameFileId, url.QueryEscape(newName),
			apiutil.PcClientInfoSuffixParam())
	} else {
		fmt.Fprintf(fullUrl, "%s/renameFile.action?fileId=%s&destFileName=%s&%s",
			API_URL,
			renameFileId, url.QueryEscape(newName),
			apiutil.PcClientInfoSuffixParam())
	}
	httpMethod := "POST"
	dateOfGmt := apiutil.DateOfGmtStr()
	appToken := p.appToken
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppRenameFile occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(respBody))
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "FileAlreadyExists" {
				return nil, apierror.NewApiError(apierror.ApiCodeFileAlreadyExisted, "文件已存在")
			}
		}
	}
	item := &AppFileEntity{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppRenameFile parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}