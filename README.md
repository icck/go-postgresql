# Go PostgreSQL Example

This repository provides a `docker-compose.yml` to run PostgreSQL.

## Usage

Start the database container in detached mode:

```bash
docker-compose up -d
```

The database listens on `localhost:5432` and stores data in `./data`.
