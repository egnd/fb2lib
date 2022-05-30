package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/egnd/fb2lib/internal/entities"
	"github.com/labstack/echo/v4"
)

func FB2FromRemoteZip(urlPrefix, libDir string, book entities.BookInfo, // @TODO:
	server echo.Context, client *http.Client,
) error {
	req, err := http.NewRequest(http.MethodGet, entities.BuildBookURL(
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
		fmt.Sprintf(`attachment; filename="%s.fb2"`, entities.BuildBookName(book.Index)),
	)

	return server.Stream(http.StatusOK, "application/fb2", resp.Body)
}
