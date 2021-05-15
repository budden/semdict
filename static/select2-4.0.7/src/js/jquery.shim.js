/* global jQuery:false, $:false */
define(function () {
  var _$ = jQuery || $;

  if (_$ == null && console && console.error) {
    console.error(
      'Select2: Экземпляр jQuery или библиотеки, совместимой с jQuery, не был ' +
      'найден. Убедитесь, что вы включили jQuery раньше Select2 на вашей ' +
      'веб-странице.'
    );
  }

  return _$;
});
