
# extra DB for ers
sudo -u postgres dropdb -e cgrates2
sudo -u postgres createdb -e -O cgrates cgrates2
