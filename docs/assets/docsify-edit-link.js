; (function (win) {
  win.EditOnGithubPlugin = {}

  function create(docBase, docEditBase, title) {
    title = title || 'Edit on github'
    docEditBase = docEditBase || docBase.replace(/\/blob\//, '/edit/')

    function editDoc(event, vm) {
      var docName = vm.route.file

      if (docName) {
        var editLink = docEditBase + docName
        window.open(editLink)
        event.preventDefault()
        return false
      } else {
        return true
      }
    }

    win.EditOnGithubPlugin.editDoc = editDoc

    return function (hook, vm) {
      win.EditOnGithubPlugin.onClick = function (event) {
        EditOnGithubPlugin.editDoc(event, vm)
      }

      var header = [
        '<a class="github-edit-btn" title="Edit on GitHub" href="',
        docBase,
        '" target="_blank" onclick="EditOnGithubPlugin.onClick(event)">',
        'Edit on <strong>GitHub</strong>',
        '</a>'
      ].join('')

      hook.afterEach(function (html) {
        return html + header
      })
    }
  }

  win.EditOnGithubPlugin.create = create
})(window)
