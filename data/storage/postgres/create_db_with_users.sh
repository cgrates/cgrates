
#
# Sample db and users creation. Replace here with your own details
#

sudo -u postgres dropdb -e cgrates
sudo -u postgres dropuser -e cgrates
sudo -u postgres psql  -c "CREATE USER cgrates password 'CGRateS.org';" 
sudo -u postgres createdb -e -O cgrates cgrates
