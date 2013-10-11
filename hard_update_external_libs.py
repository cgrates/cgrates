#! /usr/bin/env python

import os
import os.path
from subprocess import call

libs = ('github.com/fzzy/radix/redis',
        'code.google.com/p/goconf/conf',
        'github.com/bmizerany/pq',
        'github.com/vmihailenco/msgpack',
        'github.com/ugorji/go/codec',
        'labix.org/v2/mgo',
        'github.com/cgrates/fsock',
        'github.com/go-sql-driver/mysql',
        'github.com/garyburd/redigo/redis',
        'menteslibres.net/gosexy/redis',
        'github.com/howeyc/fsnotify',
)

if __name__ == "__main__":
    go_path = os.path.join(os.environ['GOPATH'], 'src')
    for lib in libs:    
        app_dir = os.path.abspath(os.path.join(go_path,lib))
        
        if os.path.islink(app_dir): continue
        git_path = os.path.join(app_dir, '.git')
        bzr_path = os.path.join(app_dir, '.bzr')
        hg_path = os.path.join(app_dir, '.hg')
        svn_path = os.path.join(app_dir, '.svn')
        if os.path.lexists(svn_path):
            print("Updating svn %s" % app_dir)
            os.chdir(app_dir)
            call(['svn', 'update'])
        elif os.path.lexists(git_path):
            print("Updating git %s" % app_dir)
            os.chdir(app_dir)
            call(['git', 'checkout', 'master'])
            call(['git', 'pull'])
        elif os.path.lexists(bzr_path):
            print("Updating bzr %s" % app_dir)
            os.chdir(app_dir)
            call(['bzr', 'pull'])
        elif os.path.lexists(hg_path):
            print("Updating hg %s" % app_dir)
            os.chdir(app_dir)
            call(['hg', 'pull', '-uv'])
        else:
            continue
        call(['go', 'install'])
