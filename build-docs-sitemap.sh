#!/bin/sh

echo "https://docs.imgproxy.net" > docs/sitemap.txt
RE='^\* \[.+\]\((.+)\)'
grep -E "$RE" docs/_sidebar.md | sed -E "s/$RE/https:\\/\\/docs.imgproxy.net\\/\\1/" >> docs/sitemap.txt
