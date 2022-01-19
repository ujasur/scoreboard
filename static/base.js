/*!
 * Generated using the Bootstrap Customizer (https://getbootstrap.com/docs/3.4/customize/)
 */

/*!
 * Bootstrap v3.4.1 (https://getbootstrap.com/)
 * Copyright 2011-2020 Twitter, Inc.
 * Licensed under the MIT license
 */

if (typeof jQuery === 'undefined') {
  throw new Error('Bootstrap\'s JavaScript requires jQuery')
}
+function ($) {
  'use strict';
  var version = $.fn.jquery.split(' ')[0].split('.')
  if ((version[0] < 2 && version[1] < 9) || (version[0] == 1 && version[1] == 9 && version[2] < 1) || (version[0] > 3)) {
    throw new Error('Bootstrap\'s JavaScript requires jQuery version 1.9.1 or higher, but lower than version 4')
  }
}(jQuery);

/* ========================================================================
 * Bootstrap: collapse.js v3.4.1
 * https://getbootstrap.com/docs/3.4/javascript/#collapse
 * ========================================================================
 * Copyright 2011-2019 Twitter, Inc.
 * Licensed under MIT (https://github.com/twbs/bootstrap/blob/master/LICENSE)
 * ======================================================================== */

/* jshint latedef: false */

+function ($) {
  'use strict';

  // COLLAPSE PUBLIC CLASS DEFINITION
  // ================================

  var Collapse = function (element, options) {
    this.$element      = $(element)
    this.options       = $.extend({}, Collapse.DEFAULTS, options)
    this.$trigger      = $('[data-toggle="collapse"][href="#' + element.id + '"],' +
                           '[data-toggle="collapse"][data-target="#' + element.id + '"]')
    this.transitioning = null

    if (this.options.parent) {
      this.$parent = this.getParent()
    } else {
      this.addAriaAndCollapsedClass(this.$element, this.$trigger)
    }

    if (this.options.toggle) this.toggle()
  }

  Collapse.VERSION  = '3.4.1'

  Collapse.TRANSITION_DURATION = 350

  Collapse.DEFAULTS = {
    toggle: true
  }

  Collapse.prototype.dimension = function () {
    var hasWidth = this.$element.hasClass('width')
    return hasWidth ? 'width' : 'height'
  }

  Collapse.prototype.show = function () {
    if (this.transitioning || this.$element.hasClass('show')) return

    var activesData
    var actives = this.$parent && this.$parent.children('.panel').children('.show, .collapsing')

    if (actives && actives.length) {
      activesData = actives.data('bs.collapse')
      if (activesData && activesData.transitioning) return
    }

    var startEvent = $.Event('show.bs.collapse')
    this.$element.trigger(startEvent)
    if (startEvent.isDefaultPrevented()) return

    if (actives && actives.length) {
      Plugin.call(actives, 'hide')
      activesData || actives.data('bs.collapse', null)
    }

    var dimension = this.dimension()

    this.$element
      .removeClass('collapse')
      .addClass('collapsing')[dimension](0)
      .attr('aria-expanded', true)

    this.$trigger
      .removeClass('collapsed')
      .attr('aria-expanded', true)

    this.transitioning = 1

    var complete = function () {
      this.$element
        .removeClass('collapsing')
        .addClass('collapse show')[dimension]('')
      this.transitioning = 0
      this.$element
        .trigger('shown.bs.collapse')
    }

    if (!$.support.transition) return complete.call(this)

    var scrollSize = $.camelCase(['scroll', dimension].join('-'))

    this.$element
      .one('bsTransitionEnd', $.proxy(complete, this))
      .emulateTransitionEnd(Collapse.TRANSITION_DURATION)[dimension](this.$element[0][scrollSize])
  }

  Collapse.prototype.hide = function () {
    if (this.transitioning || !this.$element.hasClass('show')) return

    var startEvent = $.Event('hide.bs.collapse')
    this.$element.trigger(startEvent)
    if (startEvent.isDefaultPrevented()) return

    var dimension = this.dimension()

    this.$element[dimension](this.$element[dimension]())[0].offsetHeight

    this.$element
      .addClass('collapsing')
      .removeClass('collapse show')
      .attr('aria-expanded', false)

    this.$trigger
      .addClass('collapsed')
      .attr('aria-expanded', false)

    this.transitioning = 1

    var complete = function () {
      this.transitioning = 0
      this.$element
        .removeClass('collapsing')
        .addClass('collapse')
        .trigger('hidden.bs.collapse')
    }

    if (!$.support.transition) return complete.call(this)

    this.$element
      [dimension](0)
      .one('bsTransitionEnd', $.proxy(complete, this))
      .emulateTransitionEnd(Collapse.TRANSITION_DURATION)
  }

  Collapse.prototype.toggle = function () {
    if (this.$element.hasClass('show')) {
      this.hide();
    } else {
      this.show();
    }        
  }

  Collapse.prototype.getParent = function () {
    return $(document).find(this.options.parent)
      .find('[data-toggle="collapse"][data-parent="' + this.options.parent + '"]')
      .each($.proxy(function (i, element) {
        var $element = $(element)
        this.addAriaAndCollapsedClass(getTargetFromTrigger($element), $element)
      }, this))
      .end()
  }

  Collapse.prototype.addAriaAndCollapsedClass = function ($element, $trigger) {
    var isOpen = $element.hasClass('show')

    $element.attr('aria-expanded', isOpen)
    $trigger
      .toggleClass('collapsed', !isOpen)
      .attr('aria-expanded', isOpen)
  }

  function getTargetFromTrigger($trigger) {
    var href
    var target = $trigger.attr('data-target')
      || (href = $trigger.attr('href')) && href.replace(/.*(?=#[^\s]+$)/, '') // strip for ie7

    return $(document).find(target)
  }


  // COLLAPSE PLUGIN DEFINITION
  // ==========================

  function Plugin(option) {
    return this.each(function () {
      var $this   = $(this)
      var data    = $this.data('bs.collapse')
      var options = $.extend({}, Collapse.DEFAULTS, $this.data(), typeof option == 'object' && option)

      if (!data && options.toggle && /show|hide/.test(option)) options.toggle = false
      if (!data) $this.data('bs.collapse', (data = new Collapse(this, options)))
      if (typeof option == 'string') data[option]()
    })
  }

  var old = $.fn.collapse

  $.fn.collapse             = Plugin
  $.fn.collapse.Constructor = Collapse


  // COLLAPSE NO CONFLICT
  // ====================

  $.fn.collapse.noConflict = function () {
    $.fn.collapse = old
    return this
  }


  // COLLAPSE DATA-API
  // =================

  $(document).on('click.bs.collapse.data-api', '[data-toggle="collapse"]', function (e) {
    var $this   = $(this)

    if (!$this.attr('data-target')) e.preventDefault()

    var $target = getTargetFromTrigger($this)
    var data    = $target.data('bs.collapse')
    var option  = data ? 'toggle' : $this.data()

    Plugin.call($target, option)
  })

}(jQuery);


var _ = (function(){
  // Code is taken from https://github.com/jashkenas/underscore/ 
  function restArguments(func, startIndex) {
      startIndex = startIndex == null ? func.length - 1 : +startIndex;
      return function() {
        var length = Math.max(arguments.length - startIndex, 0),
            rest = Array(length),
            index = 0;
        for (; index < length; index++) {
          rest[index] = arguments[index + startIndex];
        }
        switch (startIndex) {
          case 0: return func.call(this, rest);
          case 1: return func.call(this, arguments[0], rest);
          case 2: return func.call(this, arguments[0], arguments[1], rest);
        }
        var args = Array(startIndex + 1);
        for (index = 0; index < startIndex; index++) {
          args[index] = arguments[index];
        }
        args[startIndex] = rest;
        return func.apply(this, args);
      };
    }

    var delay = restArguments(function(func, wait, args) {
        return setTimeout(function() {
          return func.apply(null, args);
        }, wait);
      });
    

    function debounce(func, wait, immediate) {
      var timeout, result;

      var later = function(context, args) {
        timeout = null;
        if (args) result = func.apply(context, args);
      };

      var debounced = restArguments(function(args) {
        if (timeout) clearTimeout(timeout);
        if (immediate) {
          var callNow = !timeout;
          timeout = setTimeout(later, wait);
          if (callNow) result = func.apply(this, args);
        } else {
          timeout = delay(later, wait, this, args);
        }

        return result;
      });

      debounced.cancel = function() {
        clearTimeout(timeout);
        timeout = null;
      };

      return debounced;
    }
    return {
      debounce
    };
})();

var api = {
  _noop: () => {},

  defaultErrorHandler: null,

  _failHandler(errorHandler) {
    return (res, xhr, status, x) => {
      var isJson = (res.getResponseHeader('Content-Type') || '').indexOf('json') > 0;
      var errorMsg = res.responseText || res.text || res.statusText;
      res.error = {};
      if (isJson) {
        try {
          res.error = JSON.parse(res.responseText) || {};
        }
        catch { }
      }
      res.error.error = (res.error.error || (errorMsg == "error" ? "Request failed" : errorMsg));
      var next = api.defaultErrorHandler || api._noop;
      if (errorHandler) {
        if (!errorHandler(res)) {
          return
        }
      }
      next(res);
    };
  },

  listUsers(success, error) {
    $.ajax('/users').done(success).fail(error);
  },
  
  userRemove(name, success, error) {
    var q = jQuery.param({ name }, true);
    $.post(`/users/remove?${q}`).done(success).fail(api._failHandler(error));
  },
  
  userAdd(user, success, error) {
    var q = jQuery.param(user, true);
    $.post(`/users/add?${q}`).done(success).fail(api._failHandler(error));
  },

  userAuth(creds, success, error) {
    var q = jQuery.param(creds, true);
    $.post(`/users/auth?${q}`).done(success).fail(api._failHandler(error));
  },
  
  linkRemove(id, success, error) {
    var q = jQuery.param({ id }, true);
    $.post(`/links/remove?${q}`).done(success).fail(api._failHandler(error));
  },
  
  linkAdd(link, success, error) {
    $.post('/links/add', JSON.stringify(link)).done(success).fail(api._failHandler(error));
  },

  linkList(success, error) {
    $.get('/links').done(success).fail(api._failHandler(error));
  },

  session(success, error) {
    $.ajax('/session').done(success).fail(api._failHandler(error));
  },

  sessionVote(score, success, error) {
    var q = jQuery.param({ score }, true);
    $.post(`/session/vote?${q}`).done(success).fail(api._failHandler(error));
  },

  sessionReset(success, error) {
    $.post('/session/reset').done(success).fail(api._failHandler(error));
  },

  sessionUnmask(success, error) {
    $.post('/session/unmask').done(success).fail(api._failHandler(error));
  },

  sessionOpen(voters, success, error) {
    var q = jQuery.param({ name: voters }, true);
    $.post(`/session/open?${q}`).done(success).fail(api._failHandler(error));
  },

  sessionClose(success, error) {
    $.post('/session/close').done(success).fail(api._failHandler(error));
  }  
};