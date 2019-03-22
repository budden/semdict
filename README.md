# "semdict" - E-mail based user registration in golang + postgresql

## Goal 
Something more or less realistic in terms of features. Securely stored passwords, expiring registration confirmation links sent over an E-mail and so on.

## State
Pre-pre alpha. In a bottom-up manner, I collect necessary elements. So
there is no project structure yet. 

## Elements

### Done 
- concept of database error handling
- postgres quoting - sqlx seem to work fine
- genExpiryDate (schedule an expiry of a link)
- genNonce (for registration confirmation links)
- SaltAndHashPassword (safe storing of passwords)
- run postgres as a user 
- sending e-mails
- confirm registration
- ssl locally
- deploy locally


## To do
- sane page titles (otherwise history is ugly)
- validate e-mails and passwords
- integration test
- deploy on hosting

# Possible future extensions
- fail2ban integration
- captcha
- now cleanup of timed out things is 'lazy'. Implement cleanup goroutine or postgresql service? 
- implement keepalive for the service https://www.linux.org.ru/forum/development/14883028
- one connect, pool of connections or what? (now using pool and crashing if something is wrong)

# Installation 
See [installation.md in doc directory](doc/installation.md)
