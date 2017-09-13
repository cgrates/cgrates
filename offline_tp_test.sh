./test.sh
gen=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=offline_tp'
go test github.com/cgrates/cgrates/apier/v1 -tags=offline_tp
ap1=$?

exit $gen && $ap1

