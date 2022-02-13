package cloudpan

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/tickstep/cloudpan189-api/cloudpan/apiutil"
)

const (
	API_PREVIEW_URL = "https://preview.cloud.189.cn"
)

// preview picture struct
type PictureSize struct {
	Width  uint
	Height uint
}

// return the new picturesize struct with the specified size
func NewPictureSize(width, height uint) PictureSize {
	return PictureSize{
		Width:  width,
		Height: height,
	}
}

// return the string
func (p PictureSize) String() string {
	return fmt.Sprintf("%d_%d", p.Width, p.Height)
}

// get file preview picture url
func (p *PanClient) AppGetPhotoPreviewUrl(fileId string, pictureSize PictureSize) string {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	fullUrl := new(strings.Builder)
	signature := strings.ToLower(apiutil.PreviewPhotoSignatureOfHmac(p.appToken.SessionSecret,
		p.appToken.SessionKey,
		fileId, pictureSize.String(),
		timestamp))
	fmt.Fprintf(fullUrl, "%s/image/clientImageAction?fileId=%s&size=%s&sessionKey=%s&signature=%s&timeStamp=%s",
		API_PREVIEW_URL,
		fileId,
		pictureSize.String(),
		p.appToken.SessionKey,
		signature,
		timestamp,
	)
	return fullUrl.String()
}
