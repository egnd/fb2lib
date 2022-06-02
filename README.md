# fb2lib

[![Go Reference](https://pkg.go.dev/badge/github.com/egnd/fb2lib.svg)](https://pkg.go.dev/github.com/egnd/fb2lib)
[![Go Report Card](https://goreportcard.com/badge/github.com/egnd/fb2lib)](https://goreportcard.com/report/github.com/egnd/fb2lib)
[![Coverage](https://gocover.io/_badge/github.com/egnd/fb2lib)](https://gocover.io/github.com/egnd/fb2lib)
[![Pipeline](https://github.com/egnd/fb2lib/actions/workflows/latest.yml/badge.svg)](https://github.com/egnd/fb2lib/actions?query=workflow%3ALatest)

This is a server for indexing and searching fb2-books at zip archives.

### Quick start:
1. Put your archives with books into ```books``` folder

2. Create index:
```bash
docker run --rm -t --entrypoint=indexer \
  -v $(pwd)/books:/var/libs/default:ro \
  -v $(pwd)/index:/var/index/default:rw \
  -v $(pwd)/storage:/var/storage:rw \
  -v $(pwd)/logs:/var/logs:rw \
  egnd/fb2lib
```

3. Create ```docker-compose.yml```:
```yaml
version: "3.8"
services:
  app:
    image: egnd/fb2lib
    ports:
      - 80:8080
    volumes:
      - ./cfg.yml:/configs/app.override.yml:ro
      - ./libs/default:/var/libs/default:ro
      - ./index:/var/index:rw
      - ./storage:/var/storage:rw
```

4. Run server with:
```bash
docker-compose up
```

5. Server is available at http://localhost

### Hints:
* Advanced query language - https://blevesearch.com/docs/Query-String-Query/
