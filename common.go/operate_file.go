package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// CreateNotExistDir 创建不存在的文件
func CreateNotExistDir(filename string) (err error) {
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		return os.Mkdir(filename, os.ModeAppend)
	}
	return
}

// DownloadURL2File 从链接中下载内容到指定文件
func DownloadURL2File(filename, url string, limit int64) (err error) {
	fmt.Printf("download file, name=%s, url=%s\n", filename, url)
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if limit > 0 {
		resp.Body = http.MaxBytesReader(nil, resp.Body, limit)
	}

	_, err = io.Copy(out, resp.Body)
	return err
}
