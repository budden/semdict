#!/bin/bash

pgconn="postgres://$PG_USERNAME:$PG_PWD@$PG_ADDR/$PG_DB?sslmode=disable"

PSQL=psql
PSQL_ARGS="-h $PG_HOST -d $PG_DB -U $PG_USERNAME -p $PG_PORT -q -A -X"


case "$1" in
  "migrate-pg-new" )
    echo "Postgres selected"
    echo "Please enter name of new migrate file:"
    read name
    echo "Entered name: $name"

    migrate create -ext sql -dir ./migrations/postgres -seq $name
  ;;

  "migrate-pg-up" )
    echo
    echo "Migrate postgres"
    echo "Current version:"
    migrate -database $pgconn -path ./migrations/postgres version
    echo "Up:"
    migrate -database $pgconn -path ./migrations/postgres up
    echo "Postres migrations DONE."
  ;;

  * | "--help" )

    if [ "$1" != "--help" ]; then
      echo "Command '$1' does not exist."
      echo
    fi
    echo "Commands:"
    echo "- [migrate-pg-new] - Создание нового файла миграций для Postgtes."
    echo "- [migrate-pg-up] - Выполнение обновления миграций для Posrgres"
  ;;
esac
