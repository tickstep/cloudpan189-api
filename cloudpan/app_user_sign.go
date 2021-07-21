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

	fullUrl := &strings.Builder{}
	appToken := p.appToken
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	fmt.Fprintf(fullUrl, "%s/mkt/userSign.action?clientType=TELEIPHONE&version=8.9.4&model=iPhone&osFamily=iOS&osVersion=13.7&clientSn=%s",
		API_URL, apiutil.ClientSn())
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
		"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 13_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 Ecloud/8.9.4 (iPhone; " + apiutil.ClientSn() + "; appStore) iOS/13.7",
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	body, err1 := p.client.Fetch("GET", fullUrl.String(), nil, headers)
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