package test

import (
	"sync"
	"testing"

	"github.com/DLL5/tools/m3u8_download"
)

func TestM3u8Download(t *testing.T) {
	name := "test.mp4"
	url := "http://192.168.8.176:9001/api/v1/download-shared-object/aHR0cDovLzEyNy4wLjAuMTo5MDAwL2xkbC1saW5rL3BsYXkubTN1OD9YLUFtei1BbGdvcml0aG09QVdTNC1ITUFDLVNIQTI1NiZYLUFtei1DcmVkZW50aWFsPUtPTlFYUVU4MlVBMEdSUFBGMTJEJTJGMjAyNDA1MjUlMkZ1cy1lYXN0LTElMkZzMyUyRmF3czRfcmVxdWVzdCZYLUFtei1EYXRlPTIwMjQwNTI1VDE4MTYzMlomWC1BbXotRXhwaXJlcz00MzIwMCZYLUFtei1TZWN1cml0eS1Ub2tlbj1leUpoYkdjaU9pSklVelV4TWlJc0luUjVjQ0k2SWtwWFZDSjkuZXlKaFkyTmxjM05MWlhraU9pSkxUMDVSV0ZGVk9ESlZRVEJIVWxCUVJqRXlSQ0lzSW1WNGNDSTZNVGN4Tmpjd05ERTRNeXdpY0dGeVpXNTBJam9pYldsdWFXOWhaRzFwYmlKOS4wbmJDY280UXR6SnlDdDNGLW9qZjdDQ3F4bkdkNWRPQTVOeDFGb1ZnVlB1dUxVMzkwOUxsQmY3aXRUUXdQdld1Mld3ekxfZTE2MDI3STZzd0xkb2NYdyZYLUFtei1TaWduZWRIZWFkZXJzPWhvc3QmdmVyc2lvbklkPW51bGwmWC1BbXotU2lnbmF0dXJlPTgxNzJkNzFjMmY2NjVhNjk2MTIwODg5YmNkYWFiY2UxZDM5MWYwZTc4NjIwZWYzNzcxYzVkMjAyNmM5ODYwZjc="
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m := m3u8_download.New(name, url, nil)
			err := m.Run()
			if err != nil {
				t.Errorf("M3u8DownloadTest failed:%s", err.Error())
			}
		}()
	}
	wg.Wait()
}
