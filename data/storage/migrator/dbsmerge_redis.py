#!/usr/bin/python

# depends:
#   ^ redis # install via easy_install redis
# asserts:
#   ^ destination redis is not password protected when connected from source redis server
#     (https://github.com/antirez/redis/pull/2507)
# behaviour:
#   ^ the script will not overwrite keys on the destination server/database
keymask     = '*'
timeout     = 2000

import time
import redis
import os

from_host =str(os.environ["cgr_from_host"])
from_port = int(os.environ["cgr_from_port"])
from_db =int(os.environ["cgr_from_db"])
from_pass =os.environ["cgr_from_pass"]

to_host =os.environ["cgr_to_host"]
to_port =int(os.environ["cgr_to_port"])
to_db =int(os.environ["cgr_to_db"])
# to_pass =os.environ["cgr_to_pass"] # Not used

from_redis = redis.Redis(host = from_host, port = from_port, password=from_pass, db = from_db)
to_redis = redis.Redis(host = to_host, port = to_port, db = to_db)

to_keys = to_redis.keys(keymask)
from_keys = from_redis.keys(keymask)
print('Found %d keys on source.' % len(from_keys))
print('Found %d keys on destination.' % len(to_keys))

# keys found
if len(from_keys) > 0:
    # same server
    if from_host == to_host and from_port == to_port:
        print('Migrating on same server...')
        i = 0
        for key in from_keys:
            i += 1
            print('Moving key %s (%d of %d)...' % (key, i, len(from_keys)))
            from_redis.execute_command('MOVE', key, to_db)

    # different servers
    else:
        print('Migrating between different servers...')
        i = 0
        for key in from_keys:
            i += 1
            print('Moving key %s (%d of %d)...' % (key, i, len(from_keys))),
            try:
                from_redis.execute_command('MIGRATE', to_host, to_port, key, to_db, timeout)
            except redis.exceptions.ResponseError, e:
                if not 'ERR Target key name is busy' in str(e):
                    raise e
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