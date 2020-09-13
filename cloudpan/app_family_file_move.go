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


// AppFamilyMoveFile 移动文件/文件夹
func (p *PanClient) AppFamilyMoveFile(familyId int64, fileId string, destParentId string) (*AppFileEntity, *apierror.ApiError) {
	fullUrl := &strings.Builder{}

	fmt.Fprintf(fullUrl, "%s/family/file/moveFile.action?familyId=%d&fileId=%s&destFileName=%s&destParentId=%s&%s",
		API_URL, familyId, fileId, url.QueryEscape(""), destParentId, apiutil.PcClientInfoSuffixParam())
	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppFamilyMoveFile occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	item := &AppFileEntity{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppFamilyMoveFile parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}