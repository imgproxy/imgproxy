package imagedata

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
)

func upload(ctx context.Context, imageURL string, data []byte) error {
	clientCopy := *downloadClient
	client := &clientCopy
	reqCtx, reqCancel := func() (context.Context, context.CancelFunc) {
		var timeout time.Duration = time.Duration(config.DownloadTimeout) * time.Second
		return context.WithDeadline(ctx, time.Now().Add(timeout))
	}()
	req, err := http.NewRequestWithContext(reqCtx, "PUT", imageURL, bytes.NewReader(data))
	if err != nil {
		reqCancel()
		return newImageRequestError(err)
	}
	res, err := client.Do(req)
	fmt.Print(res)
	return err
}
