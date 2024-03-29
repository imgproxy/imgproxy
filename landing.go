package main

import "net/http"

var landingTmpl = []byte(`
<!doctype html>
<html>
	<head><title>Hey, I'm imgproxy!</title></head>
	<body style="background: #0d0f15">
		<style>
			* {
				color: #fff;
				font-family: Helvetica, Arial, sans-serif;
			}
			a {
				color: #40a6ff;
				text-decoration: none;
			}
			p {
				font-size:24px;
				text-align: center;
			}
		</style>
		<a href="https://imgproxy.net/" target="_blank" style="display: block; width: 266px; margin: 0 auto">
			<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="266" height="100" viewBox="0 0 1883 710">
				<path fill="#fff" d="M665 302.1v20.7h41.2v82.7H665v20.7h104.6v-20.7h-40.4V302.1H665zm34.9-31c0 10.1 5.9 17.2 17.4 17.2 11.5 0 17.4-7.1 17.4-17.2s-5.9-17.2-17.4-17.2c-11.5 0-17.4 7.1-17.4 17.2zm102.8 31v124h23v-80c0-17.8 6.1-25.8 19.7-25.8s19.7 8.1 19.7 25.8v80h23v-80c0-18 6.1-25.8 19.7-25.8s19.7 7.9 19.7 25.8v80h23v-86.4c0-24.6-14.9-40.9-37.2-40.9-11.7 0-22.4 5.2-28.5 13.6h-4.2c-6.1-8.7-15.3-13.6-26.6-13.6-11.7 0-20.5 4.1-25.3 11.6h-2.9v-8.3h-23.1zm303.2 120v-120h-23v11.2h-6.3c-5.9-8.7-19-14.5-33.1-14.5-33.1 0-55.2 24.4-55.2 61.2 0 36.2 22.2 60.4 55.2 60.4 13.6 0 26.8-5.4 33.1-13.6h6.3V420c0 19-13.4 30.2-36 30.2-18 0-31.4-7.4-36-19.8h-22.6c1.3 24.4 24.7 40.5 58.6 40.5 36.6-.1 59-19.5 59-48.8zm-94.6-62.1c0-24.6 13.8-39.7 36.4-39.7 22.6 0 36.4 15.1 36.4 39.7s-13.8 39.7-36.4 39.7c-22.6 0-36.4-15.1-36.4-39.7zm203.4-61.2c-15.1 0-28.9 6.2-34.7 15.7h-6.3v-12.4h-23v165.4h23v-53.8h6.3c5.8 9.5 19.7 15.7 34.7 15.7 33.9 0 56.5-26.1 56.5-65.3s-22.6-65.3-56.5-65.3zm-41.8 65.4c0-27.1 14.2-43.8 37.7-43.8 23.4 0 37.7 16.7 37.7 43.8 0 27.1-14.2 43.8-37.7 43.8-23.5 0-37.7-16.8-37.7-43.8zm133.8-62.1v20.7h29.4v82.7h-29.4v20.7h97.6v-20.7h-45.2V360c0-26.3 8.2-39.7 24.3-39.7 13 0 19.7 7.6 19.7 22.3v14.6h23.8v-17.9c0-25.2-13.6-40.5-36-40.5-11.9 0-21.8 4.5-27.2 12.8h-6.3v-9.5h-50.7zM1520 429.5c37 0 61.7-26.1 61.7-65.3s-24.7-65.3-61.7-65.3-61.7 26.1-61.7 65.3 24.6 65.3 61.7 65.3zm-38.7-65.3c0-27.1 14.7-43.8 38.7-43.8 24.1 0 38.7 16.7 38.7 43.8 0 27.1-14.6 43.8-38.7 43.8s-38.7-16.8-38.7-43.8zm127.6 62h26.4l29.3-43 2.3-6h2.1l2.3 6 28.9 43h28l-40.8-61 43.3-63.1H1705l-32.8 47.6-2.3 6h-2.1l-2.3-6-31.4-47.6h-27.6l42.9 65.1-40.5 59zm146.9-124.1 45 113.7h18.8l-7.5 21.1c-3.8 9.3-7.5 13.2-14.4 13.2-7.5 0-11.5-3.9-15.5-13.2h-20.9c1.7 21.7 15.3 33.9 37.7 33.9 15.1 0 27.4-9.9 33.5-26.9l50.6-141.8h-23.4l-32.8 93h-9l-36.6-93h-25.5z"/>
				<defs><path id="a" d="M0 160h659v555H0z"/></defs>
				<clipPath id="b"><use xlink:href="#a" overflow="visible"/></clipPath>
				<g clip-path="url(#b)"><linearGradient id="c" x1="277.262" x2="277.262" y1="157.87" y2="552" gradientTransform="matrix(1 0 0 -1 0 712)" gradientUnits="userSpaceOnUse"><stop offset="0" stop-color="#1d40b2"/><stop offset="1" stop-color="#1680d6"/></linearGradient><path fill="url(#c)" d="M554.5 232.4V160h-72.3v20.1H313.4V160h-72.3v20.1H72.3V160H0v72.4h20.1v88.5H0v72.4h20.1v88.5H0v72.4h72.3V534h168.8v20.1h72.3V534h168.8v20.1h72.3v-72.4h-20.1v-88.5h20.1v-72.4h-20.1v-88.5h20.1zm-297.3-56.3h40.2v40.2h-40.2v-40.2zM16.1 216.3v-40.2h40.2v40.2H16.1zm0 160.9V337h40.2v40.2H16.1zM56.3 538H16.1v-40.2h40.2V538zm241.1 0h-40.2v-40.2h40.2V538zm184.8-36.2H313.4v-20.1h-72.3v20.1H72.3v-20.1H52.2v-88.5h20.1v-72.4H52.2v-88.5h20.1v-20.1h168.8v20.1h72.3v-20.1h168.8v20.1h20.1v88.5h-20.1v72.4h20.1v88.5h-20.1v20.1zm56.3-4V538h-40.2v-40.2h40.2zm0-160.8v40.2h-40.2V337h40.2zm-40.2-120.7v-40.2h40.2v40.2h-40.2z"/><path fill="#1960c4" d="M285.3 328.9h-16.1V349h-20.1v16.1h20.1v20.1h16.1v-20.1h20.1V349h-20.1v-20.1z"/><path fill="#fff" fill-rule="evenodd" d="M575.4 651.4h77.8L519.8 517.9l55.6 133.5zm-55.6-133.5v189l55.5-55.6" clip-rule="evenodd"/></g>
				<path fill="none" d="M0 0h1883v710H0z"/>
			</svg>
		</a>
		<div style="max-width: 720px; margin: 0 auto; padding 0 15px">
			<p>
				Hey, I'm imgproxy&mdash;a fast and secure server for processing images!</a>
			</p>
			<p>
				You can get me here: <a href="https://imgproxy.net/" target="_blank">https://imgproxy.net/
			</p>
		</div>
	</body>
</html>
`)

func handleLanding(reqID string, rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "text/html")
	rw.WriteHeader(200)
	rw.Write(landingTmpl)
}
