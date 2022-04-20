# bookshelf

This is a server for indexing and searching fb2-books at zip archives.

### Quick start:
1. Put your archives with books into ```library``` folder

2. Create index:
```bash
docker run --rm -t --entrypoint=indexer \
    -v $(pwd)/index:/var/index \
    -v $(pwd)/logs:/var/logs \
    -v $(pwd)/library:/var/library \
    registry.gitlab.com/egnd/bookshelf:latest -workers=4 -batchsize=300
```

3. Create ```docker-compose.yml```:
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
```

4. Run server with:
```bash
docker-compose up
```

5. Server is available at http://localhost:8080

### Hints:
* Advanced query language - https://blevesearch.com/docs/Query-String-Query/
* Server is able to convert fb2 to epub "on-the-fly"
