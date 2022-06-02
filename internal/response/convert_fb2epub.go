package response

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func ConvertFB2Epub(converterDir string, book entities.BookInfo,
	libs entities.Libraries, server echo.Context, logger zerolog.Logger,
) error {
	filePath := book.Src
	if lib, ok := libs[book.LibName]; ok {
		filePath = path.Join(lib.Dir, filePath)
	} else {
		server.NoContent(http.StatusInternalServerError)
		return errors.New("can't define book library")
	}

	epubPath := path.Join(converterDir, strings.TrimSuffix(path.Base(book.Src), ".fb2")+".epub")
	cmd := exec.Command("bin/fb2c", "convert", "--ow", "--to=epub", filePath, converterDir)

	logger.Info().Str("epub", epubPath).Str("cmd", cmd.String()).Msg("fb2epub")

	out, err := cmd.CombinedOutput()
	if _, existsErr := os.Stat(epubPath); existsErr != nil {
		logger.Error().Str("out", string(out)).Msg("fb2c output")
	}

	if err != nil {
		return err
	}

	return server.Attachment(epubPath, entities.BuildBookName(book.Index)+".epub")
}
