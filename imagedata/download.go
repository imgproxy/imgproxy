package imagedata

var (
// Fetcher *imagefetcher.Fetcher
// downloader *imagedownloader.Downloader

// For tests
// redirectAllRequestsTo string
)

// type DownloadOptions struct {
// 	Header    http.Header
// 	CookieJar http.CookieJar
// }

// func initDownloading() error {
// 	ts, err := transport.NewTransport()
// 	if err != nil {
// 		return err
// 	}

// 	Fetcher, err = imagefetcher.NewFetcher(ts, imagefetcher.NewConfigFromEnv())
// 	if err != nil {
// 		return ierrors.Wrap(err, 0, ierrors.WithPrefix("can't create image fetcher"))
// 	}

// 	downloader = imagedownloader.NewDownloader(Fetcher)

// 	return nil
// }

// func download(ctx context.Context, imageURL string, opts DownloadOptions, secopts security.Options) (imagedatanew.ImageData, error) {
// 	// We use this for testing
// 	if len(redirectAllRequestsTo) > 0 {
// 		imageURL = redirectAllRequestsTo
// 	}

// 	return downloader.Download(ctx, imageURL, imagedownloader.DownloadOptions{
// 		Header:    opts.Header,
// 		CookieJar: opts.CookieJar,
// 	}, secopts)
// }

// func RedirectAllRequestsTo(u string) {
// 	redirectAllRequestsTo = u
// }

// func StopRedirectingRequests() {
// 	redirectAllRequestsTo = ""
// }
