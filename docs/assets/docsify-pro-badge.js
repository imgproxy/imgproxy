; (function (win) {
  win.ProBadgePlugin = {}

  function create() {
    var regex = /\!\[pro\]\((\S+)\)/g;
    var proLink = '<a class="badge" href="https://imgproxy.net/#pro" target="_blank">' +
      '<img src="$1" title="This feature is available in imgproxy Pro"/>' +
    '</a>';

    return function (hook) {
      hook.beforeEach(function (content, next) {
        content = content.replaceAll(regex, proLink)
        next(content)
      })
    }
  }

  win.ProBadgePlugin.create = create
})(window)
