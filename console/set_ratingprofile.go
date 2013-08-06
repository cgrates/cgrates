package console

import (
	"fmt"
	"github.com/cgrates/cgrates/apier/v1"
)

func init() {
	commands["set_ratingprofile"] = &CmdSetrRatingProfile{}
}

// Commander implementation
type CmdSetrRatingProfile struct {
	rpcMethod string
	rpcParams *apier.AttrSetRatingProfile
	rpcResult string
}

// name should be exec's name
func (self *CmdSetrRatingProfile) Usage(name string) string {
	return fmt.Sprintf("\n\tUsage: cgr-console [cfg_opts...{-h}] set_ratingprofile <tpid> <rateprofileid>")
}

// set param defaults
func (self *CmdSetrRatingProfile) defaults() error {
	self.rpcMethod = "ApierV1.SetRatingProfile"
	self.rpcParams = &apier.AttrSetRatingProfile{}
	return nil
}

// Parses command line args and builds CmdBalance value
func (self *CmdSetrRatingProfile) FromArgs(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf(self.Usage(""))
	}
	// Args look OK, set defaults before going further
	self.defaults()
	self.rpcParams.TPid = args[2]
	self.rpcParams.RateProfileId = args[3]
	return nil
}

func (self *CmdSetrRatingProfile) RpcMethod() string {
	return self.rpcMethod
}

func (self *CmdSetrRatingProfile) RpcParams() interface{} {
	return self.rpcParams
}

func (self *CmdSetrRatingProfile) RpcResult() interface{} {
	return &self.rpcResult
}
