#!/usr/bin/python

# depends:
#   ^ psycopg2 (debian: python-psycopg2)

import psycopg2

host = '127.0.0.1'
port = 5432
database = 'cgrates'
user = 'cgrates'
password = 'CGRateS.org'

print('Connecting to PostgreSQL...')
cnx = psycopg2.connect(
                        host=host,
                        port=port,
                        dbname=database,
                        user=user,
                        password=password
                      )
cursor = cnx.cursor()

print('Renaming old column...')
cursor.execute('ALTER TABLE cdrs RENAME COLUMN usage to usage_old')

print('Adding new column...')
cursor.execute('ALTER TABLE cdrs ADD usage NUMERIC(30)')

print('Setting new values...')
cursor.execute(
                (
                  'UPDATE cdrs SET usage = usage_old * 1000000000'
                  ' WHERE usage_old IS NOT NULL'
                )
              )

print('Deleting old column...')
cursor.execute('ALTER TABLE cdrs DROP COLUMN usage_old')

print('Commiting...')
cnx.commit()

print('Closing PostgreSQL connection...')
cnx.close()
