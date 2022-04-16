package response

import (
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

func ConvertFB2Epub(converterDir string, book entities.BookIndex, server echo.Context, logger zerolog.Logger) error {
	epubPath := path.Join(converterDir, strings.TrimSuffix(path.Base(book.Src), ".fb2")+".epub")
	cmd := exec.Command("bin/fb2c", "convert", "--ow", "--to=epub", book.Src, converterDir)

	logger.Info().Str("epub", epubPath).Str("cmd", cmd.String()).Msg("fb2epub")

	out, err := cmd.CombinedOutput()
	if _, existsErr := os.Stat(epubPath); existsErr != nil {
		logger.Error().Str("out", string(out)).Msg("fb2c output")
	}

	if err != nil {
		return err
	}

	return server.Attachment(epubPath, BuildBookName(book)+".epub")
}
