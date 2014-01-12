#! /usr/bin/env sh

FS_CDR_CSV_DIR=/var/log/freeswitch/cdr-csv
CGR_CDRC_IN_DIR=/var/log/cgrates/cdr/in/csv

/usr/bin/fs_cli -x "cdr_csv rotate"

find $FS_CDR_CSV_DIR -maxdepth 1 -mindepth 1 -not -name *.csv -print0 | xargs -0 mv -t $CGR_CDRC_IN_DIR

exit 0


