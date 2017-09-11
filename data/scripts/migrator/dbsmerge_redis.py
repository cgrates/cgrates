#!/usr/bin/python

# depends:
#   ^ redis # install via easy_install redis
# asserts:
#   ^ destination redis is not password protected when connected from source redis server
#     (https://github.com/antirez/redis/pull/2507)
# behaviour:
#   ^ the script will not overwrite keys on the destination server/database

from_db     = 11
to_db       = 10

keymask     = '*'
timeout     = 2000

import time
import redis
import argparse
parser = argparse.ArgumentParser()
parser.add_argument("-host", "--host",default="127.0.0.1", help='default: "127.0.0.1"')
parser.add_argument("-port", "--port", type=int ,default=6379, help='default: 6379')
parser.add_argument("-pass", "--password", default="", help='default: ""')

args = parser.parse_args()

from_host = args.host
from_port = args.port
from_pass = args.password

from_redis = redis.Redis(host = from_host, port = from_port, password=from_pass, db = from_db)
to_redis = redis.Redis(host = from_host, port = from_port, db = to_db)

to_keys = to_redis.keys(keymask)
from_keys = from_redis.keys(keymask)
print('Found %d keys on source.' % len(from_keys))
print('Found %d keys on destination.' % len(to_keys))

# keys found
if len(from_keys) > 0:
    print('Migrating on same server...')
    i = 0
    for key in from_keys:
        i += 1
        print('Moving key %s (%d of %d)...' % (key, i, len(from_keys)))
        from_redis.execute_command('MOVE', key, to_db)

        print('Done.')
    # done
    from_keys_after = from_redis.keys(keymask)
    to_keys_after = to_redis.keys(keymask)
    print('There are now %d keys on source.' % len(from_keys_after))
    print('There are now %d keys on destination.' % len(to_keys_after))
    print('%d keys were moved' % (len(to_keys_after) - len(to_keys)))
    print('Migration complete.')
# no keys found
else:
    print('No keys with keymask %s found in source database' % keymask)
