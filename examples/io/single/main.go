package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

func main() {
	url := "https://github.com/onivim/oni/releases/download/v0.3.4/Oni-0.3.4-amd64-linux.deb"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server return non-200 status: %s\n", resp.Status)
		return
	}

	size := resp.ContentLength

	// create dest
	destName := filepath.Base(url)
	dest, err := os.Create(destName)
	if err != nil {
		fmt.Printf("Can't create %s: %v\n", destName, err)
		return
	}
	defer dest.Close()

	p := mpb.New(mpb.WithWidth(60), mpb.WithRefreshRate(180*time.Millisecond))

	sbEta := make(chan time.Time)
	sbSpeed := make(chan time.Time)
	bar := p.AddBar(size,
		mpb.PrependDecorators(
			decor.CountersKibiByte("% 6.1f / % 6.1f", decor.WC{W: 18}),
		),
		mpb.AppendDecorators(
			decor.Name("["),
			decor.ETA(decor.ET_STYLE_MMSS, 60, sbEta),
			decor.Name("] "),
			decor.SpeedKibiByte("% .2f", 60, sbSpeed),
		),
	)

	// create proxy reader
	reader := bar.ProxyReader(resp.Body, sbEta, sbSpeed)

	// and copy from reader, ignoring errors
	io.Copy(dest, reader)

	p.Wait()
}
