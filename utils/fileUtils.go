package utils

import (
	"os"

	"pr.net/shared"
)

func CreateFile(path string, fileName string, body string) error {
	var err error
	var f *os.File
	shared.Block{
		Try: func() {
			f, err = os.Create(path + fileName)
			shared.CheckErr(err)

			_, err = f.Write([]byte(body))
			shared.CheckErr(err)

			err = f.Close()
			shared.CheckErr(err)

			err = nil
		},
		Catch: func(e shared.Exception) {
			err = e.(error)
		},
	}.Do()

	return err
}
