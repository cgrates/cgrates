#!/usr/bin/python

# depends:
#   ^ pymongo # install via: easy_install pymongo
# behaviour:
#   ^ the script will "move" the collections if source and target server are
#     the same but will "copy" (dump/restore) if source and target servers are
#     different

import subprocess
import sys
from collections import OrderedDict
from urllib import quote_plus

from pymongo import MongoClient

from_host = '127.0.0.1'
from_port = '27017'
from_db = '11'
from_auth_db = 'cgrates'  # Auth db on source server
from_user = 'cgrates'
from_pass = ''

to_host = '127.0.0.1'
to_port = '27017'
to_db = '10'
to_auth_db = "cgrates"  # Auth db on target server
to_user = 'cgrates'
to_pass = ''

ignore_empty_cols = True
# Do not migrate collections with 0 document count.
# Works only if from/to is on same host.

# Overwrite target collections flag.
# Works only if from/to is on same host.
# If from/to hosts are different we use mongorestore
# which overwrites by default.
drop_target = False

dump_folder = 'dump'

# same server
if from_host == to_host and from_port == to_port:
        print('Migrating on same server...')
        mongo_from_url = 'mongodb://%s:%s@%s:%s/%s' % (
                                                        from_user,
                                                        quote_plus(from_pass),
                                                        from_host,
                                                        from_port,
                                                        from_auth_db
                                                      )
        if from_pass == '':  # disabled auth
            mongo_from_url = 'mongodb://%s:%s/%s' % (
                                                      from_host,
                                                      from_port,
                                                      from_db
                                                    )
        client = MongoClient(mongo_from_url)

        db = client[from_db]
        cols = db.collection_names()

        # collections found
        if len(cols) > 0:
            print('Found %d collections on source. Moving...' % len(cols))
            i = 0
            for col in cols:
                i += 1
                if(
                    not ignore_empty_cols or
                    (ignore_empty_cols and db[col].count() > 0)
                  ):
                    print(
                           'Moving collection %s (%d of %d)...' % (
                             col, i, len(cols)
                           )
                         )
                    try:
                        client.admin.command(
                            OrderedDict(
                                [
                                  (
                                    'renameCollection',
                                    from_db + '.' + col
                                  ),
                                  (
                                    'to',
                                    to_db + '.' + col
                                  ),
                                  (
                                    'dropTarget',
                                    drop_target
                                  )
                                ]
                            )
                        )
                    except:
                        e = sys.exc_info()[0]
                        print(e)
                else:
                    print(
                           'Skipping empty collection %s (%d of %d)...' % (
                             col, i, len(cols)
                           )
                         )
        # no collections found
        else:
            print('No collections in source database.')

# different servers
else:
    print('Migrating between different servers...')
    print('Dumping...')
    out = subprocess.check_output([
      'mongodump',
      '--host', '%s' % from_host,
      '-u',     '%s' % from_user,
      '-p',     '%s' % from_pass,
      '--authenticationDatabase', '%s' % from_auth_db,
      '--db',     '%s' % from_db,
      '--port', '%s' % from_port,
      '-o',     '%s' % dump_folder,
      ], stderr=subprocess.STDOUT)
    print('Dump complete.')

    print('Restoring...')
    out = subprocess.check_output([
      'mongorestore',
      '--host', '%s' % to_host,
      '-u',     '%s' % to_user,
      '-p',     '%s' % to_pass,
      '--authenticationDatabase', '%s' % to_auth_db,
      '--db',   '%s' % to_db,
      '--port', '%s' % to_port,
      '--drop', '%s/%s' % (dump_folder, from_db),
      ], stderr=subprocess.STDOUT)
    print('Restore complete.')
print('Migration complete.')
