#! /usr/bin/env sh

# ./local_test.sh
# lcl=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestFreeswitchCalls'
go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestFreeswitchCalls
gnr1=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestKamailioCalls'
go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestKamailioCalls
gnr2=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestOpensipsCalls'
go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestOpensipsCalls
gnr3=$?
echo 'go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestAsteriskCalls'
go test github.com/cgrates/cgrates/general_tests -tags=call -run=TestAsteriskCalls
gnr4=$?

exit $gnr1 && $gnr2 && $gnr3 && $gnr4  