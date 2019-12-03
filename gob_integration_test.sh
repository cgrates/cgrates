#! /usr/bin/env sh
go clean --cache
./test.sh
gen=$?
echo 'go test github.com/cgrates/cgrates/apier/v1 -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v1 -tags=integration -rpc=*gob
ap1=$?
echo 'go test github.com/cgrates/cgrates/apier/v2 -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/apier/v2 -tags=integration -rpc=*gob
ap2=$?
echo 'go test github.com/cgrates/cgrates/engine  -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/engine -tags=integration -rpc=*gob
en=$?
echo 'go test github.com/cgrates/cgrates/cdrc -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/cdrc -tags=integration -rpc=*gob
cdrc=$?
echo 'go test github.com/cgrates/cgrates/ers -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/ers -tags=integration -rpc=*gob
ers=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/general_tests -tags=integration -rpc=*gob
gnr=$?
# echo 'go test github.com/cgrates/cgrates/agents -tags=integration'
# go test github.com/cgrates/cgrates/agents -tags=integration
# agts=$?
echo 'go test github.com/cgrates/cgrates/sessions -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/sessions -tags=integration -rpc=*gob
smg=$?
# echo 'go test github.com/cgrates/cgrates/dispatchers -tags=integration'
# go test github.com/cgrates/cgrates/dispatchers -tags=integration
# dis=$?
echo 'go test github.com/cgrates/cgrates/loaders -tags=integration -rpc=*gob'
go test github.com/cgrates/cgrates/loaders -tags=integration -rpc=*gob
lds=$?

exit $gen && $ap1 && $ap2 && $en && $cdrc #&& $gnr && $agts && $smg && $dis && $lds && $ers 
