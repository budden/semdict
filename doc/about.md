# What is this for?

Initially it was an attempt to create an IT explanatory dictionary with a translation option. 
Due to the lack of resources, it degenerated into just an excercise and "proof of concept" 
showing that I can write couple lines of code and which I can demonstrate to potential employers :) 

# What is here? 

There are «Sign up» functionality, cookie based stateful sessions which are stored in the database. 
Session key is used to authenticate logged in user. Sign up page sends a registration confirmation E-mail with 
a confirmation key. 

# Technology stack

- golang
- gin
- html/template
- sqlx
- postgresql
- nginx
- systemd
- VPS

# Security

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
"non-unique key" which we are capturing and handling. All other things cause server to complain to the log and then quit. 
There is a "graceful" exit (we wait couple seconds to let http server to quit and close database connection politely), and
"hard" crash where we just exit the app after only printing a notice about the error. 

To implement it we had to re-think the recommended approach to the error handling in golang. Later we found that our 
insights are very similar to the opinions expressed by "pro" golang developers, like this one: [Panic like a pro](https://hackernoon.com/panic-like-a-pro-89044d5a2d35)

We didn't really run this service in production, so it is not yet know how successful our current error handling is, 
but we believe it is a "right thing", and after some tuning it would work fine. 


