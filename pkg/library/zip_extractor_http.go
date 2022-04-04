package library

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type ZipExtractorHTTP struct {
	url    string
	client IHTTPClient
}

func FactoryZipExtractorHTTP(urlPrefix, libDir string, client IHTTPClient) IExtractorFactory {
	return func(filePath string) (IZipExtractor, error) {
		url, err := url.Parse(urlPrefix)
		if err != nil {
			return nil, err
		}

		url.Path = path.Join(url.Path, strings.TrimPrefix(filePath, libDir))

		return NewZipExtractorHTTP(url.String(), client), nil
	}
}

func NewZipExtractorHTTP(url string, client IHTTPClient) *ZipExtractorHTTP {
	return &ZipExtractorHTTP{
		url:    url,
		client: client,
	}
}

func (z *ZipExtractorHTTP) Close() error {
	return nil
}

func (z *ZipExtractorHTTP) GetSection(from int64, to int64) (res io.ReadCloser, err error) {
	var req *http.Request
	if req, err = http.NewRequest(http.MethodGet, z.url, nil); err != nil {
		return
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", from, from+to))

	var resp *http.Response
	if resp, err = z.client.Do(req); err != nil {
		return
	}

	return resp.Body, nil
}
