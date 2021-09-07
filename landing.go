package main

import "net/http"

var landingTmpl = []byte(`
<!doctype html>
<html>
	<head><title>Hey, I'm imgproxy!</title></head>
	<body>
		<h1>Hey, I'm imgproxy!</h1>
		<p style="font-size:1.2em">You can get me here: <a href="https://github.com/imgproxy/imgproxy" target="_blank">https://github.com/imgproxy/imgproxy</a></p>
	</body>
</html>
`)

func handleLanding(reqID string, rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(200)
	rw.Write(landingTmpl)
}
