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
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/library-go/logger"
)

type (
	AppUserSignStatus int
	AppUserSignResult struct {
		Status AppUserSignStatus
		Tip string
	}

	userSignResult struct {
		XMLName xml.Name `xml:"userSignResult"`
		Result int `xml:"result"`
		ResultTip string `xml:"resultTip"`
		ActivityFlag int `xml:"activityFlag"`
		PrizeListUrl string `xml:"prizeListUrl"`
		ButtonTip string `xml:"buttonTip"`
		ButtonUrl string `xml:"buttonUrl"`
		ActivityTip string `xml:"activityTip"`
	}
)

const (
	AppUserSignStatusFailed AppUserSignStatus = 0
	AppUserSignStatusSuccess AppUserSignStatus = 1
	AppUserSignStatusHasSign AppUserSignStatus = -1
)

// AppUserSign 用户签到
func (p *PanClient) AppUserSign() (*AppUserSignResult, *apierror.ApiError) {
	result := AppUserSignResult{}
	fullUrl := API_URL + "//mkt/userSign.action"
	headers := map[string]string {
		"SessionKey": p.appToken.SessionKey,
	}
	logger.Verboseln("do request url: " + fullUrl)
	body, err1 := p.client.Fetch("GET", fullUrl, nil, headers)
	if err1 != nil {
		logger.Verboseln("AppUserSign occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))
	item := &userSignResult{}
	if err := xml.Unmarshal(body, item); err != nil {
		logger.Verboseln("AppUserSign parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	switch item.Result {
	case 1:
		result.Status = AppUserSignStatusSuccess
		break
	case -1:
		result.Status = AppUserSignStatusHasSign
		break
	default:
		result.Status = AppUserSignStatusFailed
	}
	result.Tip = item.ResultTip
	return &result, nil
}