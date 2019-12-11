; (function (win) {
  win.GitterPlugin = {}

  function create(room, color) {
    color = color || "blue";

    var url = "https://gitter.im/" + room;
    var badgeUrl = "https://img.shields.io/gitter/room/" +
      room +
      "?color=" + color +
      "&style=for-the-badge" +
      "&logo=gitter";
    var html = "<a class=\"gitter-link\" href=\"" + url +"\" target=\"_blank\">" +
      "<img alt=\"Chat on Gitter\" src=\"" + badgeUrl + "\">" +
      "</a>";

    return function (hook) {
      hook.mounted(function () {
        var el = Docsify.dom.create('div', html);
        var appName = Docsify.dom.find('.app-name');

        Docsify.dom.toggleClass(el, 'gitter');
        Docsify.dom.appendTo(appName, el);
      })
    }
  }

  win.GitterPlugin.create = create
})(window)
