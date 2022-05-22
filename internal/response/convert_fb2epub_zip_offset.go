package response

import (
	"compress/flate"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func ConvertFB2EpubZipOffset(converterDir string, book entities.BookInfo,
	libs entities.Libraries, server echo.Context, logger zerolog.Logger,
) error {
	epubPath := path.Join(converterDir, book.Index.ID+".epub")
	if _, err := os.Stat(epubPath); err == nil {
		return server.File(epubPath)
	}

	fb2Path := path.Join(converterDir, book.Index.ID+".fb2")
	if _, err := os.Stat(fb2Path); err != nil {
		archivePath := strings.Split(book.Src, ".zip")[0] + ".zip"
		if lib, ok := libs[book.LibName]; ok {
			archivePath = path.Join(lib.Dir, archivePath)
		} else {
			server.NoContent(http.StatusInternalServerError)
			return errors.New("can't define book library")
		}

		zipFile, err := os.Open(archivePath)
		if err != nil {
			return err
		}
		defer zipFile.Close()

		fb2Stream := flate.NewReader(io.NewSectionReader(zipFile, int64(book.Offset), int64(book.SizeCompressed)))
		defer fb2Stream.Close()

		tmpFB2File, err := os.Create(fb2Path)
		if err != nil {
			return err
		}
		defer tmpFB2File.Close()

		if _, err := io.Copy(tmpFB2File, fb2Stream); err != nil {
			return err
		}
	}
	defer os.Remove(fb2Path)

	cmd := exec.Command("bin/fb2c", "convert", "--to=epub", fb2Path, converterDir)

	logger.Info().Str("fb2", fb2Path).Str("epub", epubPath).Str("cmd", cmd.String()).Msg("fb2epub zip offset")

	out, err := cmd.CombinedOutput()
	if _, existsErr := os.Stat(epubPath); existsErr != nil {
		logger.Error().Str("out", string(out)).Msg("fb2c output")
	}

	if err != nil {
		return err
	}

	return server.Attachment(epubPath, BuildBookName(book.Index)+".epub")
}
