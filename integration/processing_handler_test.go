package integration

// import (
// 	"io"
// 	"os"
// 	"path/filepath"

// 	"github.com/imgproxy/imgproxy/v3/config"
// 	"github.com/imgproxy/imgproxy/v3/instance"
// 	"github.com/imgproxy/imgproxy/v3/server"
// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/suite"
// )

// type ProcessingHandlerTestSuite struct {
// 	suite.Suite

// 	router   *server.Router
// 	instance *instance.Instance
// }

// func (s *ProcessingHandlerTestSuite) SetupSuite() {
// 	config.Reset()

// 	wd, err := os.Getwd()
// 	s.Require().NoError(err)

// 	s.T().Setenv("IMGPROXY_LOCAL_FILESYSTEM_ROOT", filepath.Join(wd, "/testdata"))
// 	s.T().Setenv("IMGPROXY_CLIENT_KEEP_ALIVE_TIMEOUT", "0")

// 	// We don't need config.LocalFileSystemRoot anymore as it is used
// 	// only during initialization
// 	config.Reset()
// 	config.AllowLoopbackSourceAddresses = true

// 	//err = initialize()
// 	s.Require().NoError(err)

// 	logrus.SetOutput(io.Discard)
// }

// func (s *ProcessingHandlerTestSuite) TeardownSuite() {
// 	//shutdown()
// 	logrus.SetOutput(os.Stdout)
// }
