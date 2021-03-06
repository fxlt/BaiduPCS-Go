package downloader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/requester/rio"
	"os"
	"time"
)

// DoDownload 执行下载
func DoDownload(durl string, savePath string, cfg *Config) {
	var (
		file rio.WriteCloserAt
		err  error
	)

	if savePath != "" {
		file, err = os.Create(savePath)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	download := NewDownloader(durl, file, cfg)

	exitDownloadFunc := make(chan struct{})

	download.OnExecute(func() {
		dc := download.GetDownloadStatusChan()
		var ts string

		for {
			select {
			case v, ok := <-dc:
				if !ok { // channel 已经关闭
					return
				}

				if v.TotalSize() <= 0 {
					ts = pcsutil.ConvertFileSize(v.Downloaded(), 2)
				} else {
					ts = pcsutil.ConvertFileSize(v.TotalSize(), 2)
				}

				fmt.Printf("\r↓ %s/%s %s/s in %s ............",
					pcsutil.ConvertFileSize(v.Downloaded(), 2),
					ts,
					pcsutil.ConvertFileSize(v.SpeedsPerSecond(), 2),
					v.TimeElapsed(),
				)
			}
		}
	})

	download.OnFinish(func() {
		close(exitDownloadFunc)
	})

	go func() {
		for {
			select {
			case <-exitDownloadFunc:
				return
			default:
				download.PrintAllWorkers()
			}
			time.Sleep(1e9)
		}
	}()
	download.Execute()
	<-exitDownloadFunc
}
