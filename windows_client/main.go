package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/getlantern/systray"
	"github.com/r3labs/sse/v2"
	"golang.design/x/clipboard"
)

//go:embed icon.ico
var icon []byte

var lastWritten int64

func main() {
	systray.Run(onReady, nil)
}

func onReady() {
	systray.SetTemplateIcon(icon, icon)
	systray.SetTitle("uniclip")
	systray.SetTooltip("uniclip")

	mEditCfg := systray.AddMenuItem("编辑配置", "打开配置文件")
	go func() {
		for range mEditCfg.ClickedCh {
			openConfig()
		}
	}()

	mQuitOrig := systray.AddMenuItem("退出", "退出uniclip")
	go func() {
		<-mQuitOrig.ClickedCh
		systray.Quit()
	}()

	go watchClipboard()
	go watchServer()
}

func watchClipboard() {
	ctx := context.Background()

	textCh := clipboard.Watch(ctx, clipboard.FmtText)
	imageCh := clipboard.Watch(ctx, clipboard.FmtImage)

	for {
		select {
		case b := <-textCh:
			sendToServer(b, "text/plain")
		case b := <-imageCh:
			sendToServer(b, "image/png")
		}
	}
}

func watchServer() {
	client := sse.NewClient(config.URL + "/watch")
	client.Subscribe("messages", func(msg *sse.Event) {
		if string(msg.ID) != systemID {
			res, err := http.Get(config.URL + "/data")
			if err != nil {
				fmt.Println("get clipboard data error:", err)
				return
			}
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				fmt.Println("get clipboard data with status", res.StatusCode, res.Status)
				return
			}
			var buf bytes.Buffer
			io.Copy(&buf, res.Body)
			contentType := res.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "text/") {
				clipboardWrite(clipboard.FmtText, buf.Bytes())
			} else if strings.HasPrefix(contentType, "image/") {
				if contentType == "image/png" {
					clipboardWrite(clipboard.FmtImage, buf.Bytes())
				} else {
					img, err := parseImageWithType(buf.Bytes(), contentType)
					if err != nil {
						fmt.Println("parse", contentType, "error:", err)
						return
					}
					var pngBuf bytes.Buffer
					err = png.Encode(&pngBuf, img)
					if err != nil {
						fmt.Println("encode png error:", err)
						return
					}
					clipboardWrite(clipboard.FmtImage, pngBuf.Bytes())
				}
			} else {
				fmt.Println("unsupported content type:", contentType)
			}
		}
	})
}

func sendToServer(b []byte, mime string) {
	// check and bypass content which written from server just now
	if time.Now().Unix()-atomic.LoadInt64(&lastWritten) < 3 {
		return
	}

	r := bytes.NewBuffer(b)
	sURL, err := buildPostURL()
	if err != nil {
		fmt.Println("build post url error:", err)
		return
	}
	res, err := http.Post(sURL, mime, r)
	if err != nil {
		fmt.Println("send to server error:", err)
		return
	}
	defer res.Body.Close()
	io.Copy(io.Discard, res.Body)
}

func buildPostURL() (string, error) {
	u, err := url.Parse(config.URL + "/data")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("id", systemID)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func parseImageWithType(b []byte, contentType string) (image.Image, error) {
	r := bytes.NewBuffer(b)
	switch contentType {
	case "image/jpg":
		fallthrough
	case "image/jpeg":
		return jpeg.Decode(r)
	case "image/gif":
		return gif.Decode(r)
	default:
		return nil, fmt.Errorf("unsupported image type: %s", contentType)
	}
}

func clipboardWrite(fmt clipboard.Format, b []byte) {
	clipboard.Write(fmt, b)
	atomic.StoreInt64(&lastWritten, time.Now().Unix())
}

func openConfig() {
	cmd := exec.Command("cmd", "/C", "start", mustGetConfigFilePath())
	if err := cmd.Run(); err != nil {
		fmt.Println("open config file error:", err)
	}
}
