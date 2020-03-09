#! /usr/bin/env sh

host="localhost"

DIR="/scripts"

mysql -u root -pCGRateS.org -h $host < "$DIR"/create_db_with_users.sql
mysql -u root -pCGRateS.org -h $host -D cgrates < "$DIR"/create_cdrs_tables.sql
mysql -u root -pCGRateS.org -h $host -D cgrates < "$DIR"/create_tariffplan_tables.sql

