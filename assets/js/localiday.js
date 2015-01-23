if (!document.querySelectorAll) {
  document.querySelectorAll = function(selectors) {
    var style = document.createElement('style'), elements = [], element;
    document.documentElement.firstChild.appendChild(style);
    document._qsa = [];

    style.styleSheet.cssText = selectors
    + '{x-qsa:expression(document._qsa && document._qsa.push(this))}';
    window.scrollBy(0, 0);
    style.parentNode.removeChild(style);

    while (document._qsa.length) {
      element = document._qsa.shift();
      element.style.removeAttribute('x-qsa');
      elements.push(element);
    }
    document._qsa = null;
    return elements;
  };
}

if (!document.querySelector) {
  document.querySelector = function(selectors) {
    var elements = document.querySelectorAll(selectors);
    return (elements.length) ? elements[0] : null;
  };
}

if (typeof (console) === 'undefined') {
  console = {
    log : function() {
    }
  };
}

var createTokenCookie = function(token, tokenType) {
  jQuery.cookie('loctoken', token, {expires: 1, path: '/'});
  jQuery.cookie('loctokentype', tokenType, {expires: 1, path: '/'});
};

var removeTokenCookie = function() {
  jQuery.cookie('loctoken', null, {path: '/'});
  jQuery.cookie('loctokentype', null, {path: '/'});
};

var getTokenCookie = function(location) {
  var searchToken = location ? location.search().token : null;
  return {
    token: searchToken || jQuery.cookie('loctoken'),
    tokenType: 'Bearer'
  };
};

window.getBrowserLocation = function(callback) {
  if (navigator.geolocation) {
    navigator.geolocation.getCurrentPosition(callback);
  }
};

(function($) {
  $(document).ready(function() {
    $('#masthead').click(function() {
      location.hash = "#!/home";
    });
    $('.nav-menu-icon').click(function() {
      var menu = $('.footer-links');
      if (menu.is(':hidden')) {
        menu.show();
      } else {
        menu.removeAttr('style');
      }
    });
    $('button').each(function() {
      $(this).mousedown(function(e) {
        $(this).addClass('pressed');
      });
      $(this).mouseup(function() {
        $(this).removeClass('pressed');
      })
    });
  });
})(jQuery);
