# bookshelf

This is a full text search server for books in your local library.
* supports fb2 files inside zip archives
* can convert fb2 to epub

### Quick start:
1. Create ```docker-compose.yml```:
```yaml
version: "3.8"
services:
  app:
    image: registry.gitlab.com/egnd/bookshelf:latest
    ports:
      - 8080:8080
    volumes:
      - ./index:/var/index:rw
      - ./library:/var/library:rw
      - ./logs:/var/logs:rw
```

2. Run it with:
```bash
docker-compose up
```

1. Put archives with books into ```./library```

2. Run indexer (2 parallel threads):
```bash
docker-compose exec app indexer -workers 2
```

5. Stop current composed instance and run it again with:
```bash
docker-compose up
```

6. Server is available at http://localhost:8080
