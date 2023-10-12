let clink = null;

if (window.DOCSIFY_ROUTER_MODE === "history") {
  clink = Docsify.dom.create("link");
  clink.rel = "canonical";
  Docsify.dom.appendTo(Docsify.dom.head, clink);
}

const documentTitleBase = document.title;

const linksMenu = '<div class="links-menu">' +
  '<a href="https://imgproxy.net" target="_blank" title="Website"><img src="/assets/website.svg" /></a>' +
  '<a href="https://github.com/imgproxy" target="_blank" title="GitHub"><img src="/assets/github.svg" /></a>' +
  '<a href="https://twitter.com/imgproxy_net" target="_blank" title="Twitter"><img src="/assets/twitter.svg" /></a>' +
  '<a href="https://discord.gg/5GgpXgtC9u" target="_blank" title="Discord"><img src="/assets/discord.svg" /></a>' +
  '<a href="https://github.com/sponsors/imgproxy" target="_blank" title="Sponsor"><img src="/assets/heart.svg" /></a>' +
  '</div>';

const docEditBase = 'https://github.com/imgproxy/imgproxy/edit/master/docs/';

const proBadge = Docsify.dom.create("img");
proBadge.setAttribute("src", "/assets/pro.svg");
proBadge.setAttribute("title", "This feature is available in imgproxy Pro");

const proBadgeRegex = /\!\[pro\]\((\S+)\)/g;
const proLink = `<a class="badge" href="https://imgproxy.net/#pro" target="_blank">${proBadge.outerHTML}</a>`;

const oldProBadge = "<i class='badge badge-pro'></i>";

const configRegex = /^\* `([^`]+)`:/gm;

const copyCodeBtn = '<button class="copy-code" title="Copy code"></button>';

const defaultVersions = [["latest", "latest"]];

const configureDocsify = (additionalVersions, latestVersion, latestTag) => {
  const versions = defaultVersions.concat(additionalVersions);

  const versionAliases = {};

  const versionSelect = ['<div class="sidebar-version-select"><select id="version-selector" name="version">'];
  versions.forEach(([version, tag]) => {
    const value = version == latestVersion ? "" : version;
    versionSelect.push(`<option value="${value}">${version}</value>`);

    if (version !== "latest") {
      versionAliases[`/${version}/(.*)`] =
        `https://raw.githubusercontent.com/imgproxy/imgproxy/${tag}/docs/$1`;
      versionAliases[`/${version}/`] =
        `https://raw.githubusercontent.com/imgproxy/imgproxy/${tag}/docs/README.md`;
    }
  });
  versionSelect.push('</select></div>');

  if (latestTag === "latest") latestTag = "master";

  window.$docsify = {
    name: '<a id="home-link" class="app-name-link" href="/"><img src="/assets/logo.svg"></a>' +
      linksMenu +
      versionSelect.join(""),
    nameLink: false,
    loadSidebar: true,
    relativePath: true,
    subMaxLevel: 3,
    auto2top: true,
    routerMode: window.DOCSIFY_ROUTER_MODE || "hash",
    noEmoji: true,
    alias: Object.assign(versionAliases, {
      '/latest/': 'README.md',
      '/latest/(.*)': '$1',
      '/([0-9]+\.[0-9]+)/(.*)': 'https://raw.githubusercontent.com/imgproxy/imgproxy/v$1.0/docs/$2',
      '/([0-9]+\.[0-9]+)/': 'https://raw.githubusercontent.com/imgproxy/imgproxy/v$1.0/docs/README.md',
      '/(.*)': `https://raw.githubusercontent.com/imgproxy/imgproxy/${latestTag}/docs/$1`,
      '/': `https://raw.githubusercontent.com/imgproxy/imgproxy/${latestTag}/docs/README.md`,
    }),
    search: {
      namespace: 'docs-imgproxy',
      depth: 6,
      // pathNamespaces: versions.map(v => "/" + v[0]),
      pathNamespaces: /^(\/(latest|([0-9]+\.[0-9]+)))?/,
    },
    namespaces: [
      {
        id: "version",
        values: versions.map(v => v[0]),
        optional: true,
        selector: "#version-selector",
      }
    ],
    plugins: window.$docsify.plugins.concat([
      (hook, vm) => {
        window.DocsifyVM = vm;

        hook.beforeEach(() => {
          if (clink)
            clink.href = "https://docs.imgproxy.net" + vm.route.path;
        });

        hook.doneEach(() => {
          const appNameLink = Docsify.dom.find("#home-link");

          if (!appNameLink) return;

          appNameLink.href = vm.config.currentNamespace;
        });

        hook.doneEach(() => {
          if (document.title != documentTitleBase)
            document.title += " | " + documentTitleBase;
        });


        hook.afterEach(html => {
          const docName = vm.route.file.replace(
            /https\:\/\/raw.githubusercontent\.com\/(.*)\/docs\//, ''
          );

          if (!docName) return html;

          const editButton = '<a class="github-edit-btn" title="Edit on GitHub" href="' +
            docEditBase + docName +
            '" target="_blank">' +
            'Edit on <strong>GitHub</strong>' +
            '</a>';

          return html + editButton;
        })

        hook.beforeEach((content, next) => {
          content = content
            .replaceAll(proBadgeRegex, proLink)
            .replaceAll(oldProBadge, proLink);

          content = content
            .replaceAll("📝", '<i class="icon icon-note"></i>')
            .replaceAll("⚠️", '<i class="icon icon-warn"></i>')
            .replaceAll("✅", '<i class="icon icon-check"></i>')
            .replaceAll("❌", '<i class="icon icon-cross"></i>');

          if (vm.route.path.endsWith('/configuration'))
            content = content.replaceAll(configRegex, '* <code id="$1">$1</code>:');

          next(content);
        })

        hook.doneEach(() => {
          const badges = Docsify.dom.findAll(".sidebar .badge-pro");
          badges.forEach(b => { b.replaceWith(proBadge.cloneNode()) });

          // Docsify cuts off "target" sometimes
          const links = Docsify.dom.findAll("a.badge");
          links.forEach(l => { l.setAttribute("target", "_blank") });

          const codeBlocks = Docsify.dom.findAll('pre[data-lang]');
          codeBlocks.forEach(elm =>
            elm.insertAdjacentHTML('beforeend', copyCodeBtn));
        })

        hook.mounted(() => {
          const content = Docsify.dom.find('.content');

          content.addEventListener('click', function(e) {
            if (!e.target.classList.contains('copy-code'))
              return;

            const btn = e.target;
            const code = Docsify.dom.find(btn.parentNode, 'code');

            navigator.clipboard.writeText(code.innerText).then(() => {
              btn.classList.add('copy-code-success');
              setTimeout(() => {
                btn.classList.remove('copy-code-success');
              }, 1500);
            }).catch((err)  =>{
              console.log(`Can't copy code: ${err}`);

              btn.classList.add('copy-code-error');
              setTimeout(() => {
                btn.classList.remove('copy-code-error');
              }, 1500);
            });
          });
        })
      }
    ])
  }
}

const initDocsify = (versions, latestVersion, latestTag) => {
  configureDocsify(versions, latestVersion, latestTag);
  window.runDocsify();
};

const VERSIONS_KEY = "imgproxy.versions";
const VERSIONS_ETAG_KEY = "imgproxy.versions.etag";

let latestVersion = "latest";
let latestTag = "latest";

let storedVersions = [];
let storedVersionsJson = localStorage.getItem(VERSIONS_KEY);
let storedVersionsEtag = localStorage.getItem(VERSIONS_ETAG_KEY);

if (storedVersionsJson) {
  try {
    storedVersions = JSON.parse(storedVersionsJson);
  } catch {
    storedVersions = [];
  }
}

if (storedVersions?.length)
  [latestVersion, latestTag] = storedVersions[0];
else {
  // Just in case storedVersions is not an array
  storedVersions = [];
  storedVersionsEtag = null;
}

fetch(
  "https://api.github.com/repos/imgproxy/imgproxy/releases",
  {
    headers: {
      "Accept": "application/json",
      "If-None-Match": storedVersionsEtag,
    },
  },
).then((resp) => {
  if (resp.status === 304) {
    initDocsify(storedVersions, latestVersion, latestTag);
    return;
  }

  if (resp.status != 200)
    throw new Error(`Can't fetch imgproxy versions: ${resp.statusText}`);

  resp.json().then((data) => {
    const uniq = {};
    const fetchedVersions = [];

    data.forEach((release) => {
      if (release.draft || release.prerelease) return;

      var tag = release.tag_name;
      var matches = tag?.match(/^v([0-9]+\.[0-9]+)/);

      if (!matches?.length) return;

      var version = matches[1];

      if (uniq[version]) return;

      fetchedVersions.push([version, tag]);
      uniq[version] = true;
    });

    if (fetchedVersions.length)
      [latestVersion, latestTag] = fetchedVersions[0];

    localStorage.setItem(VERSIONS_KEY, JSON.stringify(fetchedVersions));
    localStorage.setItem(VERSIONS_ETAG_KEY, resp.headers.get("Etag"));

    initDocsify(fetchedVersions, latestVersion, latestTag);

  }).catch((e) => {
    initDocsify(storedVersions, latestVersion, latestTag);
    throw e;
  });
}).catch((e) => {
  initDocsify(storedVersions, latestVersion, latestTag);
  throw e;
});
