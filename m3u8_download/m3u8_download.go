package m3u8download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/DLL5/tools/common.go"
	"github.com/google/uuid"
)

// M3U8Downloader m3u8下载器
type M3U8Downloader struct {
	FileName       string   `json:"file_name"`         // 转为mp4的输出视频的名字，如果没有则默认为output.mp4，如果不是.mp4后缀则自动加此后缀
	MetaURL        string   `json:"meta_url"`          // m3u8的元文件的链接
	VideoURLs      []string `json:"video_ur_ls"`       // 从元文件中解析出来的视频链接
	DirName        string   `json:"dir_name"`          // 下载文件所保存的目录
	TmpVideoName   []string `json:"tmp_video_name"`    // 所有分片文件的名字列表
	DontTransToMP4 bool     `json:"dont_trans_to_mp4"` // 是否不转换为mp4格式，默认为false，即转换为mp4
	FfmpegCmd      string   `json:"ffmpeg_cmd"`
}

func description() {
	fmt.Println("a package for dowloading video from url which has m3u8 meta info")
}

// OptFunc m3u8下载器配置函数
type OptFunc func(m *M3U8Downloader)

// New 新建一个m3u8下载器
func New(name, url string, opts []OptFunc) *M3U8Downloader {
	description()
	if name == "" {
		name = "output.mp4"
	}
	if !strings.HasSuffix(name, ".mp4") {
		name += ".mp4"
	}
	m := &M3U8Downloader{
		FileName:     name,
		MetaURL:      url,
		VideoURLs:    make([]string, 0),
		TmpVideoName: make([]string, 0),
		FfmpegCmd:    defaultFfmpegCmd(),
	}
	for _, v := range opts {
		v(m)
	}
	return m
}

// 实现Stringer接口，定义字符串格式化%s的打印内容
func (m *M3U8Downloader) String() string {
	return fmt.Sprintf("the filename:%s,the meta_url:%s", m.FileName, m.MetaURL)
}

// Parse 解析元数据
func (m *M3U8Downloader) Parse() (err error) {
	metaInfoRes, err := http.Get(m.MetaURL)
	if err != nil {
		return
	}
	defer metaInfoRes.Body.Close()

	metaInfo, err := io.ReadAll(metaInfoRes.Body)
	if err != nil {
		return
	}
	for _, v := range strings.Split(string(metaInfo), "\n") {
		if strings.HasPrefix(v, "http") {
			m.VideoURLs = append(m.VideoURLs, v)
		}
	}
	return
}

// Download 下载从元数据解析出来的全部m3u8分片链接数据
func (m *M3U8Downloader) Download() (err error) {
	fmt.Printf("start Download, %#v\n", *m)
	cur, err := os.Getwd()
	if err != nil {
		return
	}
	tmp := "tmp"
	if err = common.CreateNotExistDir(filepath.Join(cur, tmp)); err != nil {
		return
	}
	dirname := uuid.New().String()
	dirname = filepath.Join(cur, tmp, dirname)

	err = os.Mkdir(dirname, os.ModeAppend)
	if err != nil {
		return
	}
	m.DirName = dirname
	ch := make(chan int, 10)
	wg := &sync.WaitGroup{}
	tmpFileName := make([]string, len(m.VideoURLs))
	for i, v := range m.VideoURLs {
		ch <- 0
		wg.Add(1)
		go func(index int, name, url string, wgg *sync.WaitGroup) {
			defer func() { wgg.Done(); <-ch }()
			for retry := 0; retry < 3; retry++ {
				subErr := common.DownloadURL2File(name, url, 1<<30)
				if subErr == nil {
					tmpFileName[index] = name
					break
				}
			}
		}(i, filepath.Join(dirname, strconv.Itoa(i)+".ts"), v, wg)
	}
	wg.Wait()
	m.TmpVideoName = make([]string, 0, len(tmpFileName))
	for _, v := range tmpFileName {
		if v == "" {
			continue
		}
		m.TmpVideoName = append(m.TmpVideoName, v)
	}
	return
}

// Trans2MP4 将m3u8的所有分片文件转为一个mp4文件
func (m *M3U8Downloader) Trans2MP4() (err error) {
	fmt.Printf("start trans2mp4, %s\n", m)
	// ffmpeg -i "concat:input_1.ts|input_2.ts" -c copy output.mp4
	parameters := []string{"-i", "concat:" + strings.Join(m.TmpVideoName, "|"), "-c", "copy", filepath.Join(m.DirName, m.FileName)}
	cmd := exec.Command(defaultFfmpegCmd(), parameters...)
	fmt.Println(cmd.String())
	cmd.Stderr = os.Stdout
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// defaultFfmpegCmd 返回defaultFfmpegCmd命令
func defaultFfmpegCmd() string {
	return `ffmpeg`
}

// Run 运行m3u8下载器
func (m *M3U8Downloader) Run() (err error) {
	if err = m.Parse(); err != nil {
		return
	}
	if err = m.Download(); err != nil {
		return
	}
	if !m.DontTransToMP4 {
		if err = m.Trans2MP4(); err != nil {
			return
		}
	}
	return
}

// DontTransToMP4Opt 不转换mp4文件配置
func DontTransToMP4Opt(dtMP4 bool) OptFunc {
	return func(m *M3U8Downloader) { m.DontTransToMP4 = dtMP4 }
}

// FfmpegCmdOpt 使用的ffmpeg命令名配置
func FfmpegCmdOpt(ffmpeg string) OptFunc {
	return func(m *M3U8Downloader) {
		m.FfmpegCmd = ffmpeg
	}
}
