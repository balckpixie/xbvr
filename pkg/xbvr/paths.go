package xbvr

import (
	"os"
	"path/filepath"

	"github.com/ProtonMail/go-appdir"
)

var appDir string

var imgDir string
var cacheDir string
var binDir string
var indexDir string

func initPaths() {
	appDir = appdir.New("xbvr").UserConfig()

	imgDir = filepath.Join(appDir, "imageproxy")
	cacheDir = filepath.Join(appDir, "cache")
	binDir = filepath.Join(appDir, "bin")
	indexDir = filepath.Join(appDir, "search")

	_ = os.MkdirAll(appDir, os.ModePerm)
	_ = os.MkdirAll(imgDir, os.ModePerm)
	_ = os.MkdirAll(cacheDir, os.ModePerm)
	_ = os.MkdirAll(binDir, os.ModePerm)
	_ = os.MkdirAll(indexDir, os.ModePerm)
}
