# Для чего это нужно?

Изначально это была попытка создать толковый словарь ИТ с возможностью перевода. 
Из-за нехватки ресурсов он выродился в просто упражнение и "доказательство концепции"
, показывающее, что я могу написать пару строк кода и что я могу продемонстрировать потенциальным работодателям :) 

# Что здесь? 

Существуют функции «Регистрации», сеансы с отслеживанием состояния на основе файлов cookie, которые хранятся в sddb. 
Ключ сеанса используется для аутентификации вошедшего в систему пользователя. Страница регистрации отправляет электронное письмо с подтверждением
регистрации и ключом подтверждения. 

# Technology stack

- golang
- pkg/errors
- gin
- html/template
- sqlx
- postgresql
- nginx
- systemd
- VPS
- VS Code
- git

# Безопасность

Service is currently deployed on a VPS server at the www.semantic-dict.ru. Service runs beyond nginx. 
SSL setup of nginx is tuned with the help of [this article](https://habr.com/ru/post/325230/)
and qualifies as "A+" at the https://www.ssllabs.com/ssltest/ 

Passwords are hashed and salted. Confirmation keys and session ids are generated with a cryptographic RNG. 

# Fault-tolerance

Engine runs as a systemd service. Gin tends to swallow every panic at the boundary of request handler, put 
it to the log and continue to work. For instance, if something bad happened in a database transaction, like
"failed to rollback while processing panic", is there a meaningful way to continue? With current library stack, 
"handling" this error implies just printing the message and ignoring the consequences. In particular, database connection 
will return to the connection pool in a messy state, which will obviously impact subsequent activity. 

We took more stringest approach to error handling. There is a set of "known" errors, like "bad credentials", or
"non-unique key" which we are capturing and handling. All other things cause service to complain to the log and then quit. It is up to systemd to start a new instance. There is a "graceful" exit (we wait couple seconds to let http server to quit and close database connection politely) for some dedicated "half-known" errors, and "hard" crash where we just call os.Exit() after printing a notice about the error. 

To implement it we had to re-think the recommended approach to the error handling in golang. Later we found that our 
insights are very similar to the opinions expressed by "pro" golang developers, like this one: [Panic like a pro](https://hackernoon.com/panic-like-a-pro-89044d5a2d35)

We didn't really run this service in production, so it is not yet know how successful our current error handling is, 
but we believe it is a "right thing", and after some tuning it would work fine. 


