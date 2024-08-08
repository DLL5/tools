package m3u8_download

// 注意：包名m3u8_download最好与文件夹名一致，否则在引用包时需适配
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
	"github.com/panjf2000/ants"
)

var pool, _ = ants.NewPool(10000)

// M3U8Downloader m3u8下载器
type M3U8Downloader struct {
	FileName       string   `json:"file_name"`         // 转为mp4的输出视频的名字，如果没有则默认为output.mp4，如果不是.mp4后缀则自动加此后缀
	MetaURL        string   `json:"meta_url"`          // m3u8的元文件的链接
	Domain         string   `json:"domain"`            // 域名
	VideoURLs      []string `json:"video_ur_ls"`       // 从元文件中解析出来的视频链接
	DirName        string   `json:"dir_name"`          // 下载文件所保存的目录
	TmpVideoNames  []string `json:"tmp_video_names"`   // 所有分片文件的名字列表
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
		FileName:      name,
		MetaURL:       url,
		VideoURLs:     make([]string, 0),
		TmpVideoNames: make([]string, 0),
		FfmpegCmd:     defaultFfmpegCmd(),
	}
	for _, v := range opts {
		v(m)
	}
	return m
}

func (m *M3U8Downloader) SetDomain(domain string) *M3U8Downloader {
	m.Domain = domain
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
		if strings.HasPrefix(v, "#") {
			continue
		} else if strings.HasPrefix(v, "http") {
			m.VideoURLs = append(m.VideoURLs, v)
		} else {
			m.VideoURLs = append(m.VideoURLs, m.Domain+v)
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
		name := filepath.Join(dirname, strconv.Itoa(i)+".ts")
		url := v
		pool.Submit(func() {
			defer func() { wg.Done(); <-ch }()
			for retry := 0; retry < 3; retry++ {
				subErr := common.DownloadURL2File(name, url, 1<<30)
				if subErr == nil {
					tmpFileName[i] = name
					break
				}
			}
		})
	}
	wg.Wait()
	m.TmpVideoNames = make([]string, 0, len(tmpFileName))
	for _, v := range tmpFileName {
		if v == "" {
			continue
		}
		m.TmpVideoNames = append(m.TmpVideoNames, v)
	}
	return
}

// Trans2MP4 将m3u8的所有分片文件转为一个mp4文件
func (m *M3U8Downloader) Trans2MP4() (err error) {
	fmt.Printf("start trans2mp4, %s\n", m)
	// ffmpeg -i "concat:input_1.ts|input_2.ts" -c copy output.mp4
	// 注意： 命令中的参数"concat:input_1.ts|input_2.ts"前后的双引号，在下方的参数数组中不要加上，否则运行会报错
	parameters := []string{"-i", "concat:" + strings.Join(m.TmpVideoNames, "|"), "-c", "copy", filepath.Join(m.DirName, m.FileName)}
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

// ClearDownloadDir 清理文件下载到的目录
func ClearDownloadDir(m *M3U8Downloader) (err error) {
	if m != nil {
		return common.ClearDir(m.DirName)
	}
	return
}
