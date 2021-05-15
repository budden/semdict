(function () {
  // Восстановите загрузчик Select2 AMD, чтобы его можно было использовать
  // Нужен в основном в языковых файлах, куда загрузчик не вставляется
  if (jQuery && jQuery.fn && jQuery.fn.select2 && jQuery.fn.select2.amd) {
    var S2 = jQuery.fn.select2.amd;
  }
