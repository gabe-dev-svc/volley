## Local Development
### Database

You can run a temporary postgres container to get you up and running without the headache of having postgres running using the following command:

```
docker run --name volley_db \
  -e POSTGRES_USER=volleyuser \
  -e POSTGRES_PASSWORD=volleyrocks123 \
  -e POSTGRES_DB=volley_dev \
  -p 5432:5432 \
  -d postgis/postgis:16-3.4
```

Postgres URLs follow the pattern:

```
postgresql://user:password@localhost:5432/mydb
```

From there, you set the `DATABASE_URL` environment variable like so (do NOT disable SSL mode in production):

```
export DATABASE_URL=postgresql://volleyuser:volleyrocks123@localhost:5432/volley_dev?sslmode=disable
```

Finally, run the app

```
export DATABASE_URL=postgresql://volleyuser:volleyrocks123@localhost:5432/volley_dev?sslmode=disable && \
go run ./...
```