/*!
 * Select2 <%= package.version %>
 * https://select2.github.io
 *
 * Выпущен под лицензией MIT
 * https://github.com/select2/select2/blob/master/LICENSE.md
 */
;(function (factory) {
  if (typeof define === 'function' && define.amd) {
    // AMD. Зарегистрируйте как анонимный модуль.
    define(['jquery'], factory);
  } else if (typeof module === 'object' && module.exports) {
    // Node/CommonJS
    module.exports = function (root, jQuery) {
      if (jQuery === undefined) {
        // require('jQuery') возвращает фабрику, где window создаёт экземпляр jQuery,
        // нормализуя используемые модули, требующие этот шаблон, 
        // но предоставленное окно, если оно определено,
        // является noop (так работает jquery)
        if (typeof window !== 'undefined') {
          jQuery = require('jquery');
        }
        else {
          jQuery = require('jquery')(root);
        }
      }
      factory(jQuery);
      return jQuery;
    };
  } else {
    // Браузерные глобали
    factory(jQuery);
  }
} (function (jQuery) {
  // Это необходимо для того, чтобы мы могли поймать конфигурацию загрузчика AMD и использовать её
  // Внутренний файл должен быть обернут (по `banner.start.js`) в функции, которая возвращает ссылки на загрузчик AMD.
  //
  var S2 = 
