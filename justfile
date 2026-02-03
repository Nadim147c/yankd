generate:
  sqlc generate
  sed -i \
    -e 's/WHERE datas_fts.text MATCH ?/WHERE datas_fts MATCH ?/' \
    -e 's/SELECT text FROM datas_fts/SELECT rowid, rank FROM datas_fts/' \
    -e 's/BestRank interface{}/BestRank float64/' \
    internal/db/sqlc/queries.sql.go
  go fmt internal/db/sqlc/queries.sql.go

migration-create name:
  goose sqlite3 ./tmp/histroy.db -dir ./internal/db/sqlc/migrations create "{{name}}" sql

migration-rollback:
  goose sqlite3 ./tmp/histroy.db -dir ./internal/db/sqlc/migrations reset

migration-reset:
  just migration-rollback
  just migration-up

migration-up:
  goose sqlite3 ./tmp/histroy.db -dir ./internal/db/sqlc/migrations up

migration-down:
  goose sqlite3 ./tmp/histroy.db -dir ./internal/db/sqlc/migrations down

migration-redo:
  just migration-down
  just migration-up
