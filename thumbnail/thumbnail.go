package thumbnail

import (
	"github.com/nfnt/resize"
	"image"

	"image/png"

	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/scritch007/go-tools"
	"github.com/scritch007/shareit/types"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"path/filepath"
)

type ThumbnailGenerator struct {
	config      *types.Configuration
	pending     map[int]*ThumbnailRequest
	current     int
	requestChan chan *ThumbnailRequest
}

type ThumbnailRequest struct {
	Response chan<- string
	Path     string
}

func thumbnailToBase64(path string) (string, error) {
	var b bytes.Buffer
	file, err := os.Open(path)
	if nil != err {
		return "", err
	}
	_, err = b.ReadFrom(file)
	if nil != err {
		return "", err
	}
	e64 := base64.StdEncoding

	maxEncLen := e64.EncodedLen(len(b.Bytes()))
	encBuf := make([]byte, maxEncLen)

	e64.Encode(encBuf, b.Bytes())

	return fmt.Sprintf("data:image/png;base64,%s", encBuf), nil
}

func (t *ThumbnailGenerator) thumbnailPath(imgPath string) string {
	pre := filepath.Dir(imgPath)
	return path.Join(pre, ".thumbnail", filepath.Base(imgPath))
}
func (t *ThumbnailGenerator) checkAndCreateThumbnailFolder(path string) error {
	parentFolder := filepath.Dir(t.thumbnailPath(path))
	fileInfo, err := os.Stat(parentFolder)
	if nil != err {
		if os.IsNotExist(err) {
			os.Mkdir(parentFolder, os.ModePerm)
			return nil
		}
		return err
	}
	if !fileInfo.IsDir() {
		return errors.New("Thumbnail folder couldn't be created, it already exists but as something else")
	}
	return nil
}

func (t *ThumbnailGenerator) resizeImage(path string, resp chan<- string) {
	tools.LOG_DEBUG.Println("Resizing ", path)
	file, err := os.Open(path)
	if err != nil {
		tools.LOG_ERROR.Println("Failed to open file " + err.Error())
		resp <- ""
		return
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if nil != err {
		tools.LOG_ERROR.Println("Failed to decode file " + err.Error())
		resp <- ""
		return
	}
	m := resize.Thumbnail(48, 48, img, resize.NearestNeighbor)
	err = t.checkAndCreateThumbnailFolder(path)
	if nil != err {
		resp <- ""
		return
	}
	thumbnailPath := t.thumbnailPath(path)
	out, err := os.Create(thumbnailPath)
	if err != nil {
		tools.LOG_ERROR.Println("Failed to create file " + err.Error())
		resp <- ""
		return
	}

	base64Value := ""
	err = png.Encode(out, m)
	if nil != err {
		tools.LOG_ERROR.Println("Failed to encode file " + err.Error())
		resp <- ""
		return
	}
	out.Close()
	base64Value, err = thumbnailToBase64(path + ".thumbnail")
	if nil != err {
		tools.LOG_ERROR.Println("Failed to base64 file " + err.Error())
		resp <- ""
		return
	}

	resp <- base64Value
}

func NewThumbnailGenerator(config *types.Configuration) (t *ThumbnailGenerator) {
	t = new(ThumbnailGenerator)
	t.current = 0
	t.pending = make(map[int]*ThumbnailRequest)
	t.requestChan = make(chan *ThumbnailRequest)
	go func(t *ThumbnailGenerator) {
		resizeChan := make(chan string)
		for true {
			select {
			case req := <-t.requestChan:
				tools.LOG_DEBUG.Println("Got a new request for ", req.Path)
				if _, err := os.Stat(t.thumbnailPath(req.Path)); nil == err {
					res, err := thumbnailToBase64(t.thumbnailPath(req.Path))
					if nil != err {
						req.Response <- ""
					} else {
						req.Response <- res
					}

					continue
				}
				if 0 == len(t.pending) {
					go t.resizeImage(req.Path, resizeChan)
				}
				t.pending[t.current] = req
				t.current += 1
			case value := <-resizeChan:
				currentId := t.current - len(t.pending)
				currentRequest := t.pending[currentId]
				tools.LOG_DEBUG.Println("Got a response for ", currentRequest.Path)
				currentRequest.Response <- value
				delete(t.pending, currentId)
				if 0 != len(t.pending) {
					go t.resizeImage(t.pending[currentId+1].Path, resizeChan)
				}
			}
		}
	}(t)
	return t
}

func (t *ThumbnailGenerator) GetThumbnail(request *ThumbnailRequest) {
	//TODO check if the thumbnail has already been generated
	t.requestChan <- request
}
