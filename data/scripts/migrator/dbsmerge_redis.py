#!/usr/bin/python3

# depends:
#   ^ redis # install via easy_install redis
# asserts:
#   ^ destination redis is not password protected when connected from source
#      redis server (https://github.com/antirez/redis/pull/2507)
# behaviour:
#   ^ the script will not overwrite keys on the destination server/database

import redis

from_host = '127.0.0.1'
from_port = 6379
from_db = 11
from_pass = ''

to_host = '127.0.0.1'
to_port = 6379
to_db = 10
to_pass = ''  # Not used

keymask = '*'
timeout = 2000

from_redis = redis.Redis(
                          host=from_host,
                          port=from_port,
                          password=from_pass,
                          db=from_db
                        )
to_redis = redis.Redis(
                        host=to_host,
                        port=to_port,
                        db=to_db
                      )

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
                from_redis.execute_command(
                                            'MIGRATE',
                                            to_host,
                                            to_port,
                                            key,
                                            to_db,
                                            timeout
                                          )
            except redis.exceptions.ResponseError as e:
                if 'ERR Target key name is busy' not in str(e):
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
