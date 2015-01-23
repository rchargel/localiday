package controllers

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/hoisie/web"
	"github.com/rchargel/localiday/util"
)

const (
	imagesDir     = "assets/images"
	bgImagesDir   = "assets/images/background"
	pngFile       = ".png"
	jpgFile       = ".jpg"
	pngFileFormat = "image/png"
	jpgFileFormat = "image/jpeg"
)

// ImagesController the images controller.
type ImagesController struct {
	lastModified map[string]int64
	bgImages     []string
}

// CreateImagesController creates the images controller.
func CreateImagesController() *ImagesController {
	rand.Seed(int64(time.Now().Nanosecond()))
	return &ImagesController{make(map[string]int64, 10), make([]string, 0, 0)}
}

// RenderImage renders an image to the browser.
func (c *ImagesController) RenderImage(ctx *web.Context, imagePath string) {
	w := util.NewResponseWriter(ctx)
	lm := c.getLastModified(imagePath)
	if lm > 0 {
		w.LastModified = lm
	}
	if w.IsModified() {
		file, err := os.Open(imagesDir + "/" + imagePath)
		defer file.Close()
		if err != nil {
			w.SendError(util.HTTPFileNotFoundCode, err)
		} else {
			w.Format = c.getContentType(file.Name())
			lm = time.Now().Unix()
			w.LastModified = lm
			c.lastModified[imagePath] = lm
			w.Respond(file)
		}
	}
}

// RenderBGImage renders a random background image from the background image
// directory.
func (c *ImagesController) RenderBGImage(ctx *web.Context) {
	w := util.NewResponseWriter(ctx)

	if imgs, err := c.getBGImagesList(); err != nil {
		w.SendError(util.HTTPServerErrorCode, err)
	} else {
		img := imgs[rand.Intn(len(imgs))]
		file, err := os.Open(bgImagesDir + "/" + img)
		defer file.Close()
		if err == nil {
			w.Format = c.getContentType(file.Name())
			w.Respond(file)
		} else {
			w.SendError(util.HTTPServerErrorCode, err)
		}
	}

}

func (c *ImagesController) getBGImagesList() ([]string, error) {
	if len(c.bgImages) > 0 {
		return c.bgImages, nil
	}
	files, err := ioutil.ReadDir(bgImagesDir)
	if err != nil {
		return nil, err
	}
	imgs := make([]string, len(files))
	for i, file := range files {
		imgs[i] = file.Name()
	}
	c.bgImages = imgs
	return imgs, nil
}

func (c *ImagesController) getLastModified(file string) int64 {
	if i, found := c.lastModified[file]; found {
		return i
	}
	return -1
}

func (c *ImagesController) getContentType(filename string) string {
	if strings.Contains(filename, pngFile) {
		return pngFileFormat
	}
	return jpgFileFormat
}
