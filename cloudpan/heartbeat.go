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
	"github.com/tickstep/library-go/logger"
)

type heartBeatResp struct {
	Success bool `json:"success"`
}

// Heartbeat WEB端心跳包，周期默认1分钟
func (p *PanClient) Heartbeat() bool  {
	url := WEB_URL + "/heartbeat.action"
	body, err := p.client.DoGet(url)
	if err != nil {
		logger.Verboseln("heartbeat failed")
		return false
	}
	item := &heartBeatResp{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("heartbeat response failed")
		return false
	}
	return item.Success
}
