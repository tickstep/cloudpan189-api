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
	"strings"
)

type (
	userDrawPrizeResp struct {
		ActivityId string `json:"activityId"`
		Description string `json:"description"`
		IsUsed int `json:"isUsed"`
		ListId int `json:"listId"`
		PrizeGrade int `json:"prizeGrade"`
		PrizeId string `json:"prizeId"`
		PrizeName string `json:"prizeName"`
		PrizeStatus int `json:"prizeStatus"`
		PrizeType int `json:"prizeType"`
		UseDate string `json:"useDate"`
		UserId int64 `json:"userId"`
	}

	UserDrawPrizeResult struct {
		Success bool
		Tip string
	}

	ActivityTaskId string
)

const (
	ActivitySignin ActivityTaskId = "TASK_SIGNIN"
	ActivitySignPhotos ActivityTaskId = "TASK_SIGNIN_PHOTOS"
)

// 抽奖
func (p *PanClient) UserDrawPrize(taskId ActivityTaskId) (*UserDrawPrizeResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "https://m.cloud.189.cn/v2/drawPrizeMarketDetails.action?taskId=%s&activityId=ACT_SIGNIN",
		taskId)
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		return nil, apierror.NewApiErrorWithError(err)
	}

	item := &userDrawPrizeResp{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("UserDrawPrize parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}

	result := UserDrawPrizeResult{}
	if item.PrizeStatus == 1 {
		result.Success = true
		result.Tip = item.Description
		return &result, nil
	}
	return nil, apierror.NewFailedApiError("抽奖失败")
}