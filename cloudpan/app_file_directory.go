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
	// AppGetFileInfoParam 获取文件信息参数
	AppGetFileInfoParam struct {
		// 家庭云ID
		FamilyId int64
		// FileId 文件ID，支持文件和文件夹
		FileId   string
		// FilePath 文件绝对路径，支持文件和文件夹
		FilePath string
	}

	// AppGetFileInfoResult 文件信息响应值
	AppGetFileInfoResult struct {
		//XMLName xml.Name `xml:"folderInfo"`
		FileId string `xml:"id"`
		ParentId string `xml:"parentFolderId"`
		FileName string `xml:"name"`
		CreateDate string `xml:"createDate"`
		LastOpTime string `xml:"lastOpTime"`
		Path string `xml:"path"`
		Rev string `xml:"rev"`
		ParentFolderList parentFolderListNode `xml:"parentFolderList"`
		GroupSpaceId string `xml:"groupSpaceId"`
	}
	parentFolderListNode struct {
		FolderList []appGetFolderInfoNode `xml:"folder"`
	}
	appGetFolderInfoNode struct {
		Fid string `xml:"fid"`
		Fname string `xml:"fname"`
	}

	AppOrderBy string

	// FileListParam 文件列表参数
	AppFileListParam struct {
		// 家庭云ID
		FamilyId int64
		// FileId 文件ID
		FileId string
		// OrderBy 排序字段
		OrderBy OrderBy
		// OrderSort 排序顺序
		OrderSort OrderSort
		// PageNum 页数量，从1开始
		PageNum uint
		// PageSize 页大小，默认60
		PageSize uint
	}

	// AppFileListResult 文件列表响应值
	AppFileListResult struct {
		XMLName xml.Name `xml:"listFiles"`
		LastRev string `xml:"lastRev"`
		// 总数量
		Count int `xml:"fileList>count"`
		// 文件夹列表
		FolderList AppFileList `xml:"fileList>folder"`
		// 文件列表
		FileList AppFileList `xml:"fileList>file"`
	}

	// AppFileEntity 文件/文件夹信息
	AppFileEntity struct {
		// FileId 文件ID
		FileId string `xml:"id"`
		// ParentId 父文件夹ID
		ParentId string `xml:"parentId"`
		// FileMd5 文件MD5，文件夹为空，空文件默认为 D41D8CD98F00B204E9800998ECF8427E
		FileMd5 string `xml:"md5"`
		// FileName 名称
		FileName string `xml:"name"`
		// FileSize 文件大小
		FileSize int64 `xml:"size"`
		// LastOpTime 最后修改时间
		LastOpTime string `xml:"lastOpTime"`
		// CreateDate 创建时间
		CreateDate string `xml:"createDate"`
		// 文件完整路径
		Path string `xml:"path"`
		// MediaType 媒体类型
		MediaType MediaType `xml:"mediaType"`
		// IsFolder 是否是文件夹
		IsFolder bool
		// FileCount 文件夹子文件数量，对文件夹详情有效
		SubFileCount uint `xml:"fileCount"`

		StartLabel int `xml:"startLabel"`
		FavoriteLabel int `xml:"favoriteLabel"`
		Orientation int `xml:"orientation"`
		Rev string `xml:"rev"`
		FileCata int `xml:"fileCata"`
	}
	AppFileList []*AppFileEntity
)

const (
	// AppOrderByName 文件名
	AppOrderByName AppOrderBy = "filename"
	// AppOrderBySize 大小
	AppOrderBySize AppOrderBy = "filesize"
	// AppOrderByTime 时间
	AppOrderByTime AppOrderBy = "lastOpTime"
)

// AppGetBasicFileInfo 根据文件ID或者文件绝对路径获取文件信息，支持文件和文件夹
func (p *PanClient) AppGetBasicFileInfo(param *AppGetFileInfoParam) (*AppGetFileInfoResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}

	sessionKey := ""
	sessionSecret := ""
	if param.FamilyId <= 0 {
		// 个人云
		fmt.Fprintf(fullUrl, "%s/getFolderInfo.action?folderId=%s&folderPath=%s&pathList=0&dt=3&%s",
			API_URL, param.FileId, url.QueryEscape(param.FilePath), apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.SessionKey
		sessionSecret = p.appToken.SessionSecret
	} else {
		// 家庭云
		if param.FileId == "" {
			return nil, apierror.NewFailedApiError("FileId为空")
		}
		fmt.Fprintf(fullUrl, "%s/family/file/getFolderInfo.action?familyId=%d&folderId=%s&folderPath=%s&pathList=0&%s",
			API_URL, param.FamilyId, param.FileId, url.QueryEscape(param.FilePath), apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.FamilySessionKey
		sessionSecret = p.appToken.FamilySessionSecret
	}
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
		logger.Verboseln("AppGetBasicFileInfo occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(respBody))
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "FileNotFound" {
				return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
			}
			return nil, apierror.NewFailedApiError("请求出错")
		}
	}
	item := &AppGetFileInfoResult{}
	if param.FamilyId <= 0 {
		if err := xml.Unmarshal(respBody, item); err != nil {
			logger.Verboseln("AppGetBasicFileInfo parse response failed")
			return nil, apierror.NewApiErrorWithError(err)
		}
	} else {
		type familyAppGetFileInfoResult struct {
			FileId string `xml:"id"`
			ParentId string `xml:"parentId"`
			FileName string `xml:"name"`
			CreateDate string `xml:"createDate"`
			LastOpTime string `xml:"lastOpTime"`
			Path string `xml:"path"`
			Rev string `xml:"rev"`
		}
		fitem := &familyAppGetFileInfoResult{}
		if err := xml.Unmarshal(respBody, fitem); err != nil {
			logger.Verboseln("AppGetBasicFileInfo parse response failed")
			return nil, apierror.NewApiErrorWithError(err)
		}
		item = &AppGetFileInfoResult{
			FileId: fitem.FileId,
			ParentId: fitem.ParentId,
			FileName: fitem.FileName,
			CreateDate: fitem.CreateDate,
			LastOpTime: fitem.LastOpTime,
			Rev: fitem.Rev,
		}
	}

	return item, nil
}

func getAppOrderBy(by OrderBy) AppOrderBy {
	switch by {
	case OrderByName:
		return AppOrderByName
	case OrderBySize:
		return AppOrderBySize
	case OrderByTime:
		return AppOrderByTime
	default:
		return AppOrderByName
	}
}

func NewAppFileListParam() *AppFileListParam {
	return &AppFileListParam {
		FamilyId: 0,
		FileId: "-11",
		OrderBy: OrderByName,
		OrderSort: OrderAsc,
		PageNum: 1,
		PageSize: 200,
	}
}

// AppFileList 获取文件列表
func (p *PanClient) AppFileList(param *AppFileListParam) (*AppFileListResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}

	sessionKey := ""
	sessionSecret := ""
	if param.FamilyId <= 0 {
		// 个人云
		fmt.Fprintf(fullUrl, "%s/listFiles.action?folderId=%s&recursive=0&fileType=0&iconOption=0&mediaAttr=0&orderBy=%s&descending=%t&pageNum=%d&pageSize=%d&%s",
			API_URL,
			param.FileId, getAppOrderBy(param.OrderBy), param.OrderSort == OrderDesc, param.PageNum, param.PageSize,
			apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.SessionKey
		sessionSecret = p.appToken.SessionSecret
	} else {
		// 家庭云
		fmt.Fprintf(fullUrl, "%s/family/file/listFiles.action?folderId=%s&familyId=%d&fileType=0&iconOption=0&mediaAttr=0&orderBy=%d&descending=%t&pageNum=%d&pageSize=%d&%s",
			API_URL,
			param.FileId, param.FamilyId, param.OrderBy, param.OrderSort == OrderDesc, param.PageNum, param.PageSize,
			apiutil.PcClientInfoSuffixParam())
		sessionKey = p.appToken.FamilySessionKey
		sessionSecret = p.appToken.FamilySessionSecret
	}
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
		logger.Verboseln("AppFileList occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(respBody))
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "FileNotFound" {
				return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
			}
		}
	}
	item := &AppFileListResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppFileList parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}

	if item.FolderList != nil && len(item.FolderList) > 0 {
		for _,item := range item.FolderList {
			item.IsFolder = true
		}
	}
	if item.FileList != nil && len(item.FileList) > 0 {
		for _,item := range item.FileList {
			item.IsFolder = false
		}
	}

	return item, nil
}