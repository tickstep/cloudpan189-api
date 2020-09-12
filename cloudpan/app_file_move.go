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
	AppMoveFileResult struct {
		XMLName xml.Name `xml:"fileList"`
		// 总数量
		Count int `xml:"count"`
		// 文件夹列表
		FolderList AppFileList `xml:"folder"`
		// 文件列表
		FileList AppFileList `xml:"file"`
	}
)

// AppMoveFile 移动文件/文件夹
func (p *PanClient) AppMoveFile(fileIdList []string, targetFolderId string) (*AppMoveFileResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/batchMoveFile.action?fileIdList=%s&destParentFolderId=%s&%s",
		API_URL, strings.Join(fileIdList, ";"), targetFolderId, apiutil.PcClientInfoSuffixParam())
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
		logger.Verboseln("AppMoveFile occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	item := &AppMoveFileResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppMoveFile parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}