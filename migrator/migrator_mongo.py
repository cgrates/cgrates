#!/usr/bin/python

# depends:
#   ^ pymongo
# asserts:
#   ^ destination mongo database exists and the destination user can auth on it
# behaviour:
#   ^ the script will "move" the collections if source and target server are the same
#     but will "copy" (dump/restore) if source and target servers are different

from_host   = '127.0.0.1'
from_port   = '27017'
from_db     = 'cgrates2'
from_user   = 'cgrates'
from_pass   = 'CGRateS.org'

to_host     = '127.0.0.1'
to_port     = '27017'
to_db       = 'cgrates2'
to_user     = 'cgrates'
to_pass     = 'CGRateS.org'

dump_folder = 'dump'

from pymongo import MongoClient
from urllib import quote_plus
from collections import OrderedDict

mongo_from_url = 'mongodb://' + from_user + ':' + quote_plus(from_pass) + '@'+ from_host + ':' + from_port + '/' + from_db
client = MongoClient(mongo_from_url)

db = client[from_db]
cols = db.collection_names()

# collections found
if len(cols) > 0:
    # same server
    if from_host == to_host and from_port == to_port:
            print('Migrating on same server...')
            print('Found %d collections on source. Moving...' % len(cols))
            i = 0
            for col in db.collection_names():
                i += 1
                print('Moving colection %s (%d of %d)...' % (col, i, len(cols)))
                client.admin.command(OrderedDict([('renameCollection', from_db + '.' + col), ('to', to_db + '.' + col)]))
    # different servers
    else:
        import subprocess
        import os
        import shutil

        print('Migrating between different servers...')
        print('Dumping...')
        out = subprocess.check_output([
          'mongodump',
          '--host', '%s' % from_host,
          '-u',     '%s' % from_user,
          '-p',     '%s' % from_pass,
          '-d',     '%s' % from_db,
          '--port', '%s' % from_port,
          '-o',     '%s' % dump_folder,
          ], stderr= subprocess.STDOUT)
        print('Dump complete.')

        print('Restoring...')
        out = subprocess.check_output([
          'mongorestore',
          '--host', '%s' % to_host,
          '-u',     '%s' % to_user,
          '-p',     '%s' % to_pass,
          '--authenticationDatabase', '%s' % to_db,
          '--db',   '%s' % to_db,
          '--port', '%s' % to_port,
          '--drop', '%s/%s' % (dump_folder,from_db),
          ], stderr= subprocess.STDOUT)
        print('Restore complete.')
    print('Migration complete.')
# no collections found
else:
    print('No collections in source database.')
