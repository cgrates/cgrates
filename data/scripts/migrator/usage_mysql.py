#!/usr/bin/python3

# depends:
#   ^ mysql (debian: python-mysql.connector)

import mysql.connector

host = '127.0.0.1'
port = 3306
database = 'cgrates'
user = 'root'
password = 'CGRateS.org'

config = {
  'user':     user,
  'password': password,
  'host':     host,
  'port':     port,
  'database': database,
}

print('Connecting to MySQL...')
cnx = mysql.connector.connect(**config)
cursor = cnx.cursor()

print('Renaming old column...')
cursor.execute(
                (
                  'ALTER TABLE cdrs'
                  ' CHANGE COLUMN `usage` `usage_old`'
                  ' DECIMAL(30,9)'
                )
              )

print('Adding new column...')
cursor.execute('ALTER TABLE cdrs ADD `usage` DECIMAL(30)')

print('Setting new values...')
cursor.execute(
                (
                  'UPDATE cdrs SET `usage` = `usage_old` * 1000000000'
                  ' WHERE usage_old IS NOT NULL'
                )
              )

print('Deleting old column...')
cursor.execute('ALTER TABLE cdrs DROP COLUMN usage_old')

print('Closing MySQL connection...')
cnx.close()
