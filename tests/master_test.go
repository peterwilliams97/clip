package clip_test

import (
	"github.com/unidoc/unipdf/common"
)

// init sets the logging level for all tests.
func init() {
	common.SetLogger(common.NewConsoleLogger(common.LogLevelInfo))
}
