# Семантический словарь - англо-русский словарь на основе смысла

# Быстрый старт

[DEV.md](DEV.md)

## Концепция

### Идея "семантического словаря"

Есть два отличия от обычного словаря перевода с языка A на язык B:

- существует ячейка в смысле слова, а не ячейка в слове. Слова, имеющие несколько значений, имеют несколько смыслов.
- существует множество вариантов перевода, и мы отслеживаем источники перевода. Например, Oracle и Microsoft могут использовать разные переводы некоторых значений слов на русский язык. Мы создаем диалект "Oracle" для хранения переводов Oracle и диалект "Microsoft" для хранения переводов Microsoft.

### Подробные спецификации требований (на русском языке)

[Смотри сюда](doc/тз/общее.md)

## Состояние
Пре-альфа, без развертывания. Не все функции реализованы.

## Технология

### Сделано
- концепция обработки ошибок базы данных
- цитирование postgres - sqlx, похоже, работает нормально
- genExpiryDate (запланируйте истечение срока действия ссылки)
- genNonce (ссылки для подтверждения регистрации)
- SaltAndHashPassword (безопасное хранение паролей)
- запустите postgres от имени пользователя
- отправка электронных писем
- подтвердите регистрацию
- ssl локально
- развертывание локально


## Делать
- вменяемые заголовки страниц (в противном случае история уродлива)
- проверка электронной почты и паролей
- интеграционный тест
- развертывание на хостинге

# Возможные будущие расширения
- интеграция fail2ban
- капча
- теперь очистка тайм-аута 'ленива'. Реализовать программу очистки goroutine или службу postgresql?
- реализовать функцию keepalive для службы https://www.linux.org.ru/forum/development/14883028
- одно соединение, пул соединений или что? (теперь используйте пул и сбой, если что-то не так)

# Установка
См. [installation.md в каталоге doc](doc/installation.md) 
