package etcd_test

import (
	"flag"
	"os"
	"testing"
	"time"

	updater "github.com/mwitkow/go-flagz/etcd"
	"github.com/mwitkow/go-flagz/test_etcd"

	"github.com/Sirupsen/logrus"
	etcd "github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

const (
	prefix = "/updater_test/"
)

var (
	logger = logrus.StandardLogger()
	ctxNil = context.Background()
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including assertion methods.
type UpdaterTestSuite struct {
	suite.Suite
	keys etcd.KeysAPI

	flagSet *flag.FlagSet
	updater *updater.Updater
}

// Clean up the etcd state before each test.
func (s *UpdaterTestSuite) SetupTest() {
	s.keys.Delete(ctxNil, prefix, &etcd.DeleteOptions{Dir: true, Recursive: true})
	_, err := s.keys.Set(ctxNil, prefix, "", &etcd.SetOptions{Dir: true})
	if err != nil {
		s.T().Fatalf("cannot create empty dir %v: %v", prefix, err)
	}
	s.flagSet = flag.NewFlagSet("updater_test", flag.ContinueOnError)
	s.updater, err = updater.New(s.flagSet, s.keys, prefix, &testingLog{T: s.T()})
	if err != nil {
		s.T().Fatalf("cannot create updater: %v", err)
	}
}

func (s *UpdaterTestSuite) setFlagzValue(flagzName string, value string) {
	_, err := s.keys.Set(ctxNil, prefix+flagzName, value, &etcd.SetOptions{})
	if err != nil {
		s.T().Fatalf("failed setting flagz value: %v")
	}
}

func (s *UpdaterTestSuite) getFlagzValue(flagzName string) string {
	resp, err := s.keys.Get(ctxNil, prefix+flagzName, &etcd.GetOptions{})
	if err != nil {
		s.T().Fatalf("failed getting flagz value: %v")
	}
	return resp.Node.Value
}

// Tear down the updater
func (s *UpdaterTestSuite) TearDownTest() {
	s.updater.Stop()
	time.Sleep(100 * time.Millisecond)
}

func (s *UpdaterTestSuite) Test_ErrorsOnInitialUnknownFlag() {
	s.flagSet.Int("someint", 1337, "some int usage")
	s.setFlagzValue("anotherint", "999")
	s.Require().Error(s.updater.Initialize(), "initialize should complain about unknown flag")
}

func (s *UpdaterTestSuite) Test_SetsInitialValues() {
	someInt := s.flagSet.Int("someint", 1337, "some int usage")
	someString := s.flagSet.String("somestring", "initial_value", "some int usage")
	anotherString := s.flagSet.String("anotherstring", "default_value", "some int usage")
	s.setFlagzValue("someint", "2015")
	s.setFlagzValue("somestring", "changed_value")
	s.Require().NoError(s.updater.Initialize())

	s.Require().Equal(2015, *someInt, "int flag should change value")
	s.Require().Equal("changed_value", *someString, "string flag should change value")
	s.Require().Equal("default_value", *anotherString, "anotherstring should be unchanged")
}

func (s *UpdaterTestSuite) Test_DynamicUpdate() {
	someInt := s.flagSet.Int("someint", 1337, "some int usage")
	s.Require().NoError(s.updater.Initialize())
	s.Require().NoError(s.updater.Start())
	s.Require().Equal(1337, *someInt, "int flag should not change value")
	s.setFlagzValue("someint", "2014")
	eventually(s.T(), 1*time.Second,
		assert.Equal, 2014,
		func() interface{} { return *someInt },
		"someint value should change")
	s.setFlagzValue("someint", "2015")
	eventually(s.T(), 1*time.Second,
		assert.Equal, 2015,
		func() interface{} { return *someInt },
		"someint value should change")
	s.setFlagzValue("someint", "2016")
	eventually(s.T(), 1*time.Second,
		assert.Equal, 2016,
		func() interface{} { return *someInt },
		"someint value should change")
}

func (s *UpdaterTestSuite) Test_DynamicUpdateRestoresGoodState() {
	someInt := s.flagSet.Int("someint", 1337, "some int usage")
	someFloat := s.flagSet.Float64("somefloat", 1.337, "some int usage")
	s.setFlagzValue("someint", "2015")
	s.Require().NoError(s.updater.Initialize())
	s.Require().NoError(s.updater.Start())
	s.Require().Equal(2015, *someInt, "int flag should change value")
	s.Require().Equal(1.337, *someFloat, "float flag should not change value")

	// Bad update causing a rollback.
	s.setFlagzValue("someint", "randombleh")
	eventually(s.T(), 1*time.Second,
		assert.Equal,
		"2015",
		func() interface{} {
			return s.getFlagzValue("someint")
		},
		"someint failure should revert etcd value to 2015")

	// Make sure we can continue updating.
	s.setFlagzValue("someint", "2016")
	s.setFlagzValue("somefloat", "3.14")
	eventually(s.T(), 1*time.Second,
		assert.Equal, 2016,
		func() interface{} { return *someInt },
		"someint value should change, after rolled back")
	eventually(s.T(), 1*time.Second,
		assert.Equal, 3.14,
		func() interface{} { return *someFloat },
		"somefloat value should change")

}

func TestUpdaterSuite(t *testing.T) {
	server, err := test_etcd.New(os.Stderr)
	if err != nil {
		t.Fatalf("failed starting test server: %v", err)
	}
	t.Logf("will use etcd test endpoint: %v", server.Endpoint)
	defer func() {
		server.Stop()
		t.Logf("cleaned up etcd test server")
	}()
	suite.Run(t, &UpdaterTestSuite{keys: etcd.NewKeysAPI(server.Client)})
}

type assertFunc func(T assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool
type getter func() interface{}

// eventually tries a given Assert function 5 times over the period of time.
func eventually(T *testing.T, duration time.Duration,
	af assertFunc, expected interface{}, actual getter, msgAndArgs ...interface{}) {
	increment := duration / 5
	for i := 0; i < 5; i++ {
		time.Sleep(increment)
		if af(T, expected, actual(), msgAndArgs...) {
			return
		}
	}
	T.FailNow()
}

// Abstraction that allows us to pass the *testing.T as a logger to the updater.
type testingLog struct {
	T *testing.T
}

func (tl *testingLog) Printf(format string, v ...interface{}) {
	tl.T.Logf(format+"\n", v...)
}
