# fb2lib

[![Go Reference](https://pkg.go.dev/badge/github.com/egnd/fb2lib.svg)](https://pkg.go.dev/github.com/egnd/fb2lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/egnd/fb2lib)](https://goreportcard.com/report/github.com/egnd/fb2lib)
[![Coverage](https://gocover.io/_badge/github.com/egnd/fb2lib)](https://gocover.io/github.com/egnd/fb2lib)
<!-- [![Pipeline](https://github.com/egnd/fb2lib/actions/workflows/pipeline.yml/badge.svg)](https://github.com/egnd/fb2lib/actions?query=workflow%3APipeline) -->

This is a server for indexing and searching fb2-books at zip archives.

### Quick start:
1. Put your archives with books into ```library``` folder

2. Create index:
```bash
docker run --rm -t --entrypoint=indexer \
    -v $(pwd)/index:/var/index \
    -v $(pwd)/logs:/var/logs \
    -v $(pwd)/library:/var/library \
    egnd/fb2lib -workers=4 -batchsize=300
```

3. Create ```docker-compose.yml```:
```yaml
version: "3.8"
services:
  app:
    image: egnd/fb2lib
    ports:
      - 8080:8080
    volumes:
      - ./index:/var/index:rw
      - ./library:/var/library:rw
```

4. Run server with:
```bash
docker-compose up
```

5. Server is available at http://localhost:8080

### Hints:
* Advanced query language - https://blevesearch.com/docs/Query-String-Query/
* Server is able to convert fb2 to epub "on-the-fly"
