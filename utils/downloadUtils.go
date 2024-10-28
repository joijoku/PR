package utils

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"pr.net/shared"
)

func DownloadFileFromURL(fPath string, url string) (bool, string, error) {
	status := false
	msg := ""
	var err error

	shared.Block{
		Try: func() {
			// resp, err := http.Get(url)
			// shared.CheckErr(err)

			req, err := http.NewRequest("GET", url, nil)
			shared.CheckErr(err)

			// fileName := path.Base(resp.Request.URL.Path)
			fileName := path.Base(req.URL.Path)

			tr := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}

			client := &http.Client{Transport: tr}
			resp, err := client.Do(req)
			shared.CheckErr(err)

			defer resp.Body.Close()

			for _, v := range strings.Split(resp.Header.Get("Content-Disposition"), ";") {
				if strings.Contains(v, "filename") {
					fileName = strings.Replace(v, "filename=", "", 1)
				}
			}

			out, err := os.Create(fPath + fileName)
			shared.CheckErr(err)

			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			shared.CheckErr(err)

			status = true
			msg = fileName
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
			msg = e.(error).Error()
		},
	}.Do()

	return status, msg, err
}
