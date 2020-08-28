# cloudpan189-api
GO语言封装的 cloud 189 天翼云盘接口API。可以基于该接口库实现对天翼云盘的二次开发。

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/tickstep/cloudpan189-api?tab=doc)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/modern-go/concurrent/master/LICENSE)

# 快速使用

导入包
```
import "github.com/tickstep/cloudpan189-api/cloudpan"
```

先调用登录接口，获取APP端cookie和WEB端cookie
```
	appToken, e := cloudpan.AppLogin("193xxxxxx@189.cn", "123xxxxx")
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
```

使用获取到的cookie创建PanClient实例
```
	// pan client
	panClient := cloudpan.NewPanClient(*webToken, *appToken)
```

调用PanClient相关方法可以实现对cloud189云盘的相关操作
```
	// do get file info action
	fi, err1 := panClient.FileInfoByPath("/我的文档")
	if err1 != nil {
		fmt.Println("get file info error")
		return
	}
	fmt.Printf("name = %s, size = %d, path = %s", fi.FileName, fi.FileSize, fi.Path)
```