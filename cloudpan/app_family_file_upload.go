package cloudpan

import (
	"encoding/xml"
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-api/cloudpan/apiutil"
	"github.com/tickstep/library-go/logger"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
)

func (p *PanClient) AppFamilyCreateUploadFile(param *AppCreateUploadFileParam) (*AppCreateUploadFileResult, *apierror.ApiError) {
	if param.ParentFolderId == "-11" {
		param.ParentFolderId = ""
	}
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/family/file/createFamilyFile.action?fileMd5=%s&fileName=%s&familyId=%d&parentId=%s&resumePolicy=1&fileSize=%d&%s",
		API_URL, param.Md5, url.QueryEscape(param.FileName), param.FamilyId, param.ParentFolderId, param.Size,
		apiutil.PcClientInfoSuffixParam())

	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := apiutil.XRequestId()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": requestId,
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	body, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppFamilyCreateUploadFile occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))

	item := &AppCreateUploadFileResult{}
	if err := xml.Unmarshal(body, item); err != nil {
		logger.Verboseln("AppFamilyCreateUploadFile parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item.XRequestId = requestId
	return item, nil
}

func (p *PanClient) AppFamilyUploadFileData(familyId int64, uploadUrl, uploadFileId, xRequestId string, fileRange *AppFileUploadRange, uploadFunc UploadFunc) *apierror.ApiError {
	fullUrl := uploadUrl + "?" + apiutil.PcClientInfoSuffixParam()
	httpMethod := "PUT"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := xRequestId
	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	headers := map[string]string {
		"Accept": "*/*",
		"FamilyId": strconv.FormatInt(familyId, 10),
		"Content-Type": "application/octet-stream",
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl, dateOfGmt),
		"X-Request-ID": requestId,
		"ResumePolicy": "1",
		"UploadFileId": uploadFileId,
		"Edrive-UploadFileRange": "bytes=" + strconv.FormatInt(fileRange.Offset, 10) + "-" + strconv.FormatInt(fileRange.Len, 10),
		"Expect": "100-continue",
	}

	logger.Verboseln("do request url: " + fullUrl)
	resp, err1 := uploadFunc(httpMethod, fullUrl, headers)
	if err1 != nil {
		logger.Verboseln("AppUploadFileData occurs error: ", err1.Error())
		return apierror.NewApiErrorWithError(err1)
	}
	if resp != nil {
		er := &apierror.AppErrorXmlResp{}
		d, _ := ioutil.ReadAll(resp.Body)
		if err := xml.Unmarshal(d, er); err == nil {
			if er.Code != "" {
				if er.Code == "UploadOffsetVerifyFailed" {
					return apierror.NewApiError(apierror.ApiCodeUploadOffsetVerifyFailed, "上传文件数据偏移值校验失败")
				}
				return apierror.NewFailedApiError(er.Message)
			}
		}
	}
	return nil
}

func (p *PanClient) AppFamilyUploadFileCommit(familyId int64, uploadCommitUrl, uploadFileId, xRequestId string) (*AppUploadFileCommitResult, *apierror.ApiError) {
	fullUrl := uploadCommitUrl + "?" + apiutil.PcClientInfoSuffixParam()

	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := xRequestId
	headers := map[string]string {
		"FamilyId": strconv.FormatInt(familyId, 10),
		"ResumePolicy": "1",
		"uploadFileId": uploadFileId,

		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl, dateOfGmt),
		"X-Request-ID": requestId,
	}

	logger.Verboseln("do request url: " + fullUrl)
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl, nil, headers)
	if err1 != nil {
		logger.Verboseln("AppFamilyUploadFileCommit occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "UploadFileStatusVerifyFailed" {
				return nil, apierror.NewApiError(apierror.ApiCodeUploadFileStatusVerifyFailed, "上传文件校验失败")
			}
		}
	}
	item := &AppUploadFileCommitResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppFamilyUploadFileCommit parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

// AppFamilyGetUploadFileStatus 查询上传的文件状态
func (p *PanClient) AppFamilyGetUploadFileStatus(familyId int64, uploadFileId string) (*AppGetUploadFileStatusResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/family/file/getFamilyFileStatus.action?familyId=%d&uploadFileId=%s&resumePolicy=1&%s",
		API_URL, familyId, uploadFileId,
		apiutil.PcClientInfoSuffixParam())

	sessionKey := p.appToken.FamilySessionKey
	sessionSecret := p.appToken.FamilySessionSecret
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := apiutil.XRequestId()
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": sessionKey,
		"Signature": apiutil.SignatureOfHmac(sessionSecret, sessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": requestId,
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppGetUploadFileStatus occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "UploadFileNotFound" {
				return nil, apierror.NewApiError(apierror.ApiCodeUploadFileNotFound, "服务器上传文件不存在")
			}
		}
	}

	type appGetUploadFileStatusResult struct {
		XMLName xml.Name `xml:"uploadFile"`
		// 上传文件的ID
		UploadFileId string `xml:"uploadFileId"`
		// 已上传的大小
		Size int64 `xml:"dataSize"`
		FileUploadUrl string `xml:"fileUploadUrl"`
		FileCommitUrl string `xml:"fileCommitUrl"`
		FileDataExists int `xml:"fileDataExists"`
	}
	itemInternal := &appGetUploadFileStatusResult{}
	if err := xml.Unmarshal(respBody, itemInternal); err != nil {
		logger.Verboseln("AppGetUploadFileStatus parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return &AppGetUploadFileStatusResult{
		XMLName: itemInternal.XMLName,
		UploadFileId: itemInternal.UploadFileId,
		Size: itemInternal.Size,
		FileUploadUrl: itemInternal.FileUploadUrl,
		FileCommitUrl: itemInternal.FileCommitUrl,
		FileDataExists: itemInternal.FileDataExists,
	}, nil
}