package etcd


import (
	"testing"
	"github.com/stretchr/testify/suite"

	"github.com/coreos/go-etcd/etcd"
	flagz_etcd "github.com/mwitkow-io/go-flagz/etcd"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including assertion methods.
type ExampleTestSuite struct {
	suite.Suite
	VariableThatShouldStartAtFive int
}




// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (suite *ExampleTestSuite) SetupTest() {
	suite.VariableThatShouldStartAtFive = 5
}

// All methods that begin with "Test" are run as tests within a
// suite.
func (suite *ExampleTestSuite) TestExample() {
	suite.Equal(suite.VariableThatShouldStartAtFive, 5)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUpdaterSuite(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}

