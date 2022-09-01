if (window.DOCSIFY_ROUTER_MODE === "history") {
  var clink = document.createElement("link")
  clink.rel = "canonical"
  document.getElementsByTagName("head")[0].appendChild(clink)
}

var documentTitleBase = document.title;

var gitterURL = "https://gitter.im/imgproxy/imgproxy";
var gitterBadgeURL = "https://img.shields.io/gitter/room/imgproxy/imgproxy" +
  "?color=1775d3&style=for-the-badge&logo=gitter";
var gitterBadge = '<div class="gitter">' +
  '<a class="gitter-link" href="' + gitterURL + '" target="_blank">' +
  '<img alt="Chat on Gitter" src="' + gitterBadgeURL + '">' +
  '</div></a>';

var docEditBase = 'https://github.com/imgproxy/imgproxy/edit/master/docs/';

var proBadge = document.createElement("img")
proBadge.setAttribute("src", "/assets/pro.svg")
proBadge.setAttribute("title", "This feature is available in imgproxy Pro")

var proBadgeRegex = /\!\[pro\]\((\S+)\)/g;
var proLink = '<a class="badge" href="https://imgproxy.net/#pro" target="_blank">' +
  proBadge.outerHTML + '</a>';

var oldProBadge = "<i class='badge badge-pro'></i>";

var versions = ["latest"].concat(window.IMGPROXY_VERSIONS);
var latestVersion = window.IMGPROXY_VERSIONS[0];
var versionSelect = '<select id="version-selector" name="version" class="sidebar-version-select">';
versions.forEach(function (version) {
  var value = version == latestVersion ? "" : version;
  versionSelect = versionSelect + '<option value="' + value + '">' + version + '</value>';
});
versionSelect = versionSelect + '</select>';

window.$docsify = {
  name: '<a id="home-link" class="app-name-link" href="/"><img src="/assets/logo.svg"></a>' +
    gitterBadge +
    versionSelect,
  nameLink: false,
  repo: 'https://github.com/imgproxy',
  loadSidebar: true,
  relativePath: true,
  subMaxLevel: 2,
  auto2top: true,
  routerMode: window.DOCSIFY_ROUTER_MODE || "hash",
  noEmoji: true,
  alias: {
    '/latest/': 'README.md',
    '/latest/(.*)': '$1',
    '/([0-9]+\.[0-9]+)/(.*)': 'https://raw.githubusercontent.com/imgproxy/imgproxy/v$1.0/docs/$2',
    '/([0-9]+\.[0-9]+)/': 'https://raw.githubusercontent.com/imgproxy/imgproxy/v$1.0/docs/README.md',
    '/(.*)': 'https://raw.githubusercontent.com/imgproxy/imgproxy/v' + latestVersion + '.0/docs/$1',
    '/': 'https://raw.githubusercontent.com/imgproxy/imgproxy/v' + latestVersion + '.0/docs/README.md',
  },
  search: {
    namespace: 'docs-imgproxy',
    depth: 6,
    pathNamespaces: versions.map(function (v) { return "/" + v })
  },
  namespaces: [
    {
      id: "version",
      values: versions,
      optional: true,
      selector: "#version-selector"
    }
  ],
  plugins: [
    function (hook, vm) {
      window.DocsifyVM = vm
      hook.beforeEach(function () {
        if (clink) {
          clink.href = "https://docs.imgproxy.net" + vm.route.path
        }
      });

      hook.doneEach(function () {
        var appNameLink = Docsify.dom.find("#home-link");

        if (!appNameLink) return;

        appNameLink.href = vm.config.currentNamespace;
      });

      hook.doneEach(function() {
        if (document.title != documentTitleBase)
          document.title += " | " + documentTitleBase;
      });


      hook.afterEach(function (html) {
        var docName = vm.route.file.replace(
          /https\:\/\/raw.githubusercontent\.com\/(.*)\/docs\//, ''
        )

        if (!docName) {
          return html;
        }

        var editButton = '<a class="github-edit-btn" title="Edit on GitHub" href="' +
          docEditBase + docName +
          '" target="_blank">' +
          'Edit on <strong>GitHub</strong>' +
          '</a>';

        return html + editButton
      })

      hook.beforeEach(function (content, next) {
        content = content.replaceAll(proBadgeRegex, proLink)
        content = content.replaceAll(oldProBadge, proLink)
        next(content)
      })

      hook.doneEach(function () {
        var badges = Docsify.dom.findAll(".sidebar .badge-pro")
        badges.forEach(function (b) { b.replaceWith(proBadge.cloneNode()) })

        // Docsify cuts off "target" sometimes
        var links = Docsify.dom.findAll("a.badge")
        links.forEach(function(l){ l.setAttribute("target", "_blank") })
      })
    }
  ]
}
