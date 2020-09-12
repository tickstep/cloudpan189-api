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
	AppCopyFileParam struct {
		FileId string
		DestFileName string
		DestFolderId string
	}
)

// AppCopyFile 复制文件到目标文件夹
func (p *PanClient) AppCopyFile(param *AppCopyFileParam) (*AppFileEntity, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/copyFile.action?fileId=%s&destFileName=%s&destParentFolderId=%s&%s",
		API_URL,
		param.FileId, url.QueryEscape(param.DestFileName), param.DestFolderId,
		apiutil.PcClientInfoSuffixParam())
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
		logger.Verboseln("AppCopyFile occurs error: ", err1.Error())
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
		logger.Verboseln("AppCopyFile parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}