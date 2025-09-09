package integration

import (
	"context"
	"net"

	"github.com/imgproxy/imgproxy/v3"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
}

// StartImgproxy starts imgproxy instance for the tests
// Returns instance, instance address and stop function
func (s *Suite) StartImgproxy(c *imgproxy.Config) (net.Addr, context.CancelFunc) {
	ctx, cancel := context.WithCancel(s.T().Context())

	c.Server.Bind = ":0"
	c.Server.LogMemStats = true

	i, err := imgproxy.New(ctx, c)
	s.Require().NoError(err)

	addrCh := make(chan net.Addr)

	go func() {
		err = i.StartServer(s.T().Context(), addrCh)
		if err != nil {
			s.T().Errorf("Imgproxy stopped with error: %v", err)
		}
	}()

	return <-addrCh, cancel
}
