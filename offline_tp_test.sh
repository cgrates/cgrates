./test.sh
gen=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline_TP'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline_TP

exit $gen
