; (function (win) {
  win.EditOnGithubPlugin = {}

  function create(docBase) {
    var docEditBase = docBase.replace(/\/blob\//, '/edit/')

    return function (hook, vm) {
      hook.afterEach(function (html) {
        var url = docBase
        var docName = vm.route.file

        if (docName) {
          url = docEditBase + docName
        }

        var header = [
          '<a class="github-edit-btn" title="Edit on GitHub" href="',
          url,
          '" target="_blank">',
          'Edit on <strong>GitHub</strong>',
          '</a>'
        ].join('')

        return html + header
      })
    }
  }

  win.EditOnGithubPlugin.create = create
})(window)
