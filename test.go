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

package main

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/library-go/jsonhelper"
	"os"
)

type (
	userpw struct {
		UserName string `json:"username"`
		Password string `json:"password"`
	}
)

func objToJsonStr(v interface{}) string {
	r,_ := jsoniter.MarshalToString(v)
	return string(r)
}

func main() {
	configFile, err := os.OpenFile("userpw.txt", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		fmt.Println("read user info error")
		return
	}
	defer configFile.Close()

	userpw := &userpw{}
	err = jsonhelper.UnmarshalData(configFile, userpw)
	if err != nil {
		fmt.Println("read user info error")
		return
	}

	// do login
	appToken, e := cloudpan.AppLogin(userpw.UserName, userpw.Password)
	if e != nil {
		fmt.Println(e)
		return
	}

	webToken := &cloudpan.WebLoginToken{}
	webTokenStr := cloudpan.RefreshCookieToken(appToken.SessionKey)
	if webTokenStr != "" {
		webToken.CookieLoginUser = webTokenStr
	}
	fmt.Println("login success")

	// pan client
	panClient := cloudpan.NewPanClient(*webToken, *appToken)

	// do get file info action
	fi, err1 := panClient.FileInfoByPath("/我的文档")
	if err1 != nil {
		fmt.Println("get file info error")
		return
	}
	fmt.Printf("name = %s, size = %d, path = %s", fi.FileName, fi.FileSize, fi.Path)

	// get family cloud list
	ffl, err2 := panClient.AppFamilyGetFamilyList()
	if err2 != nil {
		fmt.Println("get family list error: " + err2.Error())
		return
	}
	fmt.Println(objToJsonStr(ffl))
}
