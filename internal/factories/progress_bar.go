package factories

import (
	"os"

	"github.com/schollz/progressbar/v3"
)

func NewFileProgressBar(filePath, description string) (pb *progressbar.ProgressBar, err error) {
	var fi os.FileInfo
	if fi, err = os.Stat(filePath); err != nil {
		return
	}

	pb = progressbar.DefaultBytes(fi.Size(), description)

	return
}
