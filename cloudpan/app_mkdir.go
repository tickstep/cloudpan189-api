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
		XMLName xml.Name `xml:"folder"`
		// fileId 文件ID
		FileId string `xml:"id"`
		// ParentId 父文件夹ID
		ParentId string `xml:"parentId"`
		// FileName 名称
		FileName string `xml:"name"`
		// LastOpTime 最后修改时间
		LastOpTime string `xml:"lastOpTime"`
		// CreateDate 创建时间
		CreateDate string `xml:"createDate"`
		Rev string `xml:"rev"`
		FileCata int `xml:"fileCata"`
	}
)

func (p *PanClient) AppMkdir(parentFileId, dirName string) (*AppMkdirResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/createFolder.action?parentFolderId=%s&folderName=%s&relativePath=&%s",
		API_URL, parentFileId, url.QueryEscape(dirName), apiutil.PcClientInfoSuffixParam())
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

