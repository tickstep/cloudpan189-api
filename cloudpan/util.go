package cloudpan

import (
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-api/cloudpan/lib/escaper"
	"github.com/tickstep/library-go/logger"
	"path"
	"strings"
)

const (
	// ShellPatternCharacters 通配符字符串
	ShellPatternCharacters = "*?[]"
)

func (p *PanClient) recurseMatchPathByShellPattern(familyId int64, index int, pathSlice *[]string, parentFileInfo *AppFileEntity, resultList *AppFileList) {
	if parentFileInfo == nil {
		// default root "/" entity
		parentFileInfo = NewAppFileEntityForRootDir()
		if index == 0 && len(*pathSlice) == 1 {
			// root path "/"
			*resultList = append(*resultList, parentFileInfo)
			return
		}
		p.recurseMatchPathByShellPattern(familyId, index+1, pathSlice, parentFileInfo, resultList)
		return
	}

	if index >= len(*pathSlice) {
		// 已经是最后的路径分片了，是命中的结果
		*resultList = append(*resultList, parentFileInfo)
		return
	}

	//if !strings.ContainsAny((*pathSlice)[index], ShellPatternCharacters) {
	//	// 不包含通配符，先查缓存
	//	curPathStr := path.Clean(parentFileInfo.Path + "/" + (*pathSlice)[index])
	//
	//	// try cache
	//	if v := p.loadFilePathFromCache(familyId, curPathStr); v != nil {
	//		p.recurseMatchPathByShellPattern(familyId, index+1, pathSlice, v, resultList)
	//		return
	//	}
	//}

	// 遍历目录下所有文件
	if !parentFileInfo.IsFolder {
		return
	}
	fileListParam := NewAppFileListParam()
	fileListParam.FamilyId = familyId
	fileListParam.FileId = parentFileInfo.FileId
	fileResult, err := p.AppGetAllFileList(fileListParam)
	if err != nil {
		logger.Verbosef("获取目录文件列表错误")
		return
	}
	if fileResult == nil || len(fileResult.FileList) == 0 {
		// 文件目录下文件为空
		return
	}

	curParentPathStr := parentFileInfo.Path
	if curParentPathStr == "/" {
		curParentPathStr = ""
	}

	// 先检测是否满足文件名全量匹配
	for _, fileEntity := range fileResult.FileList {
		// cache item
		fileEntity.Path = curParentPathStr + "/" + fileEntity.FileName
		//p.storeFilePathToCache(familyId, fileEntity.Path, fileEntity)

		// 云盘文件名支持*?[]等特殊符号，先排除文件名完全一致匹配的情况，这种情况下不能开启通配符匹配
		if fileEntity.FileName == (*pathSlice)[index] {
			// 匹配一个就直接返回
			p.recurseMatchPathByShellPattern(familyId, index+1, pathSlice, fileEntity, resultList)
			return
		}
	}

	// 使用通配符匹配
	for _, fileEntity := range fileResult.FileList {
		// cache item
		fileEntity.Path = curParentPathStr + "/" + fileEntity.FileName
		//p.storeFilePathToCache(familyId, fileEntity.Path, fileEntity)

		// 使用通配符
		if matched, _ := path.Match((*pathSlice)[index], fileEntity.FileName); matched {
			p.recurseMatchPathByShellPattern(familyId, index+1, pathSlice, fileEntity, resultList)
		}
	}
}

// MatchPathByShellPattern 通配符匹配文件路径, pattern为绝对路径，符合的路径文件存放在resultList中
func (p *PanClient) MatchPathByShellPattern(familyId int64, pattern string) (resultList *AppFileList, error *apierror.ApiError) {
	errInfo := apierror.NewApiError(apierror.ApiCodeFailed, "")
	resultList = &AppFileList{}

	patternSlice := strings.Split(escaper.Escape(path.Clean(pattern), []rune{'['}), PathSeparator) // 转义中括号
	if patternSlice[0] != "" {
		errInfo.Err = "路径不是绝对路径"
		return nil, errInfo
	}
	defer func() { // 捕获异常
		if err := recover(); err != nil {
			resultList = nil
			errInfo.Err = "查询路径异常"
		}
	}()

	parentFile := NewAppFileEntityForRootDir()
	if path.Clean(strings.TrimSpace(pattern)) == "/" {
		*resultList = append(*resultList, parentFile)
		return resultList, nil
	}
	p.recurseMatchPathByShellPattern(familyId, 1, &patternSlice, parentFile, resultList)
	return resultList, nil
}
