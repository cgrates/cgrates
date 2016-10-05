#! /usr/bin/env sh

USER="cgrates"
PASSWORD="CGRateS.org"

MONGO_VERSION=$(mongo --version | sed 's/.* \([0-9]\.[0-9\]\).*/\1/')
MONGO_MAJOR=$(echo $MONGO_VERSION | cut -d '.' -f 1)
MONGO_MINOR=$(echo $MONGO_VERSION | cut -d '.' -f 2)

# up to mongo 2.6 create user is done by addUser
if [ $MONGO_MAJOR -eq 2 -a $MONGO_MINOR -lt 6 ]; then
	FUNC=addUser
	ROLES="[ 'userAdminAnyDatabase' ]"
else
	FUNC=createUser
	ROLES="[{ role: 'userAdminAnyDatabase', db: 'admin' }]"
fi
mongo --quiet --eval "
db = db.getSiblingDB('admin');
db.$FUNC(
  {
    user: '$USER',
    pwd: '$PASSWORD',
    roles: $ROLES
  }
)"
cu=$?

if [ $cu = 0 ]; then
	echo ""
	echo "\t+++ CGR-DB successfully set-up! +++"
	echo ""
	exit 0
fi


