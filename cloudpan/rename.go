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

func (p *PanClient) Rename(renameFileId, newName string) (bool, *apierror.ApiError) {
	if renameFileId == "" {
		return false, apierror.NewFailedApiError("请指定命名的文件")
	}

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/renameFile.action?fileId=%s&fileName=%s",
		WEB_URL, renameFileId, url.QueryEscape(newName))
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("Rename failed")
		return false, apierror.NewApiErrorWithError(err)
	}
	comResp := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, comResp); err == nil {
		if comResp.ErrorCode == "FileAlreadyExists" {
			logger.Verboseln("Rename response failed")
			return false, apierror.NewFailedApiError("文件名已存在")
		}
	}

	result := &apierror.SuccessResp{}
	if err := json.Unmarshal(body, result); err != nil {
		logger.Verboseln("Rename response failed")
		return false, apierror.NewApiErrorWithError(err)
	}
	return result.Success, nil
}
