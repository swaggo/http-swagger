// Swagger UI Version: 4.15.5

'use strict';

function _typeof(obj) {
  if (typeof Symbol === 'function' && typeof Symbol.iterator === 'symbol') {
    _typeof = function _typeof(obj) {
      return typeof obj;
    };
  } else {
    _typeof = function _typeof(obj) {
      return obj &&
        typeof Symbol === 'function' &&
        obj.constructor === Symbol &&
        obj !== Symbol.prototype
        ? 'symbol'
        : typeof obj;
    };
  }
  return _typeof(obj);
}

// From: https://raw.githubusercontent.com/chilts/umd-template/master/template.js
(function (f) {
  // module name and requires
  var name = 'MyPlugin';
  var requires = []; // CommonJS

  if (
    (typeof exports === 'undefined' ? 'undefined' : _typeof(exports)) === 'object' &&
    typeof module !== 'undefined'
  ) {
    module.exports = f.apply(
      null,
      requires.map(function (r) {
        return require(r);
      })
    ); // RequireJS
  } else if (typeof define === 'function' && define.amd) {
    define(requires, f); // <script>
  } else {
    var g;

    if (typeof window !== 'undefined') {
      g = window;
    } else if (typeof global !== 'undefined') {
      g = global;
    } else if (typeof self !== 'undefined') {
      g = self;
    } else {
      // works providing we're not in "use strict";
      // needed for Java 8 Nashorn
      // seee https://github.com/facebook/react/issues/3037
      g = this;
    }

    g[name] = f.apply(
      null,
      requires.map(function (r) {
        return g[r];
      })
    );
  }
})(function () {
  return function (system) {
    return {
        wrapComponents: {
            info: (Original, system) => (props) => {
                const html = htm.bind(React.createElement);
                // Uses https://github.com/developit/htm to render JSX.
                return html`
                <div>
                    <h3>Hello world! I am above the Info component.</h3>
                    <Original {...props} />
                </div>`;
            }
        },
    };
  };
});
