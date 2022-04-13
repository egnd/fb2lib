package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gitlab.com/egnd/bookshelf/internal/entities"
)

func FB2FromRemoteZip(urlPrefix, libDir string, book entities.BookIndex,
	server echo.Context, client *http.Client,
) error {
	req, err := http.NewRequest(http.MethodGet, BuildBookURL(
		strings.Split(book.Src, ".zip")[0]+".zip", urlPrefix, libDir,
	), nil)
	if err != nil {
		return err
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", uint64(book.Offset), uint64(book.Offset+book.SizeCompressed)))

	resp, err := client.Do(req.WithContext(server.Request().Context()))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("remote zip response error: %s", resp.Status)
	}

	server.Response().Header().Set(echo.HeaderContentEncoding, "deflate")
	server.Response().Header().Set(echo.HeaderContentDisposition,
		fmt.Sprintf(`attachment; filename="%s.fb2"`, book.ID),
	)

	return server.Stream(http.StatusOK, "application/fb2", resp.Body)
}