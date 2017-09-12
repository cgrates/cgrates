#!/usr/bin/python

# depends:
#   ^ pymongo # install via: easy_install pymongo
# behaviour:
#   ^ the script will "move" the collections if source and target server are the same
#     but will "copy" (dump/restore) if source and target servers are different

ignore_empty_cols = True
# Do not migrate collections with 0 document count.
# Works only if from/to is on same host.

# Overwrite target collections flag.
# Works only if from/to is on same host.
# If from/to hosts are different we use mongorestore which overwrites by default.
drop_target = False

dump_folder = 'dump'
import os
import sys
from pymongo import MongoClient
from urllib import quote_plus
from collections import OrderedDict

from_host =os.environ["cgr_from_host"]
from_port =os.environ["cgr_from_port"]
from_db  =os.environ["cgr_from_db"]
from_auth_db =os.environ["cgr_from_auth_db"]
from_user =os.environ["cgr_from_user"]
from_pass =os.environ["cgr_from_pass"]

to_host =os.environ["cgr_to_host"]
to_port =os.environ["cgr_to_port"]
to_db =os.environ["cgr_to_db"]
to_auth_db =os.environ["cgr_to_auth_db"]
to_user =os.environ["cgr_to_user"]
to_pass =os.environ["cgr_to_pass"]

# same server
if from_host == to_host and from_port == to_port:
        print('Migrating on same server...')
        mongo_from_url = 'mongodb://' + from_user + ':' + quote_plus(from_pass) + '@'+ from_host + ':' + from_port + '/' + from_auth_db
        if from_pass == '': # disabled auth
          mongo_from_url = 'mongodb://' + from_host + ':' + from_port + '/' + from_db
        client = MongoClient(mongo_from_url)

        db = client[from_db]
        cols = db.collection_names()

        # collections found
        if len(cols) > 0:
            print('Found %d collections on source. Moving...' % len(cols))
            i = 0
            for col in cols:
                i += 1
                if not ignore_empty_cols or (ignore_empty_cols and db[col].count() > 0):
                    print('Moving collection %s (%d of %d)...' % (col, i, len(cols)))
                    try:
                        client.admin.command(OrderedDict([('renameCollection', from_db + '.' + col), ('to', to_db + '.' + col), ('dropTarget', drop_target)]))
                    except:
                        e = sys.exc_info()[0]
                        print(e)
                else:
                    print('Skipping empty collection %s (%d of %d)...' % (col, i, len(cols)))
        # no collections found
        else:
            print('No collections in source database.')

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
      '--authenticationDatabase', '%s' % from_auth_db,
      '--db',     '%s' % from_db,
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
      '--authenticationDatabase', '%s' % to_auth_db,
      '--db',   '%s' % to_db,
      '--port', '%s' % to_port,
      '--drop', '%s/%s' % (dump_folder, from_db),
      ], stderr= subprocess.STDOUT)
    print('Restore complete.')
print('Migration complete.')
