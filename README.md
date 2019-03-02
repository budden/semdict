# "a" - E-mail based user registration in golang + postgresql

## Goal 
Something more or less realistic in terms of features. Securely stored passwords, expiring registration confirmation links sent over an E-mail and so on.

## State
Pre-pre alpha. In a bottom-up manner, I collect necessary elements. So
there is no project structure yet. 

## Elements

### Done or all clear
- concept of database error handling
- postgres quoting - sqlx seem to work fine
- genExpiryDate (schedule an expiry of a link)
- genNonce (for registration confirmation links)
- hashAndSaltPassword (safe storing of passwords)
- run postgres as self (non-root) - done once, but steps were not recorded very well


## To do
- ssl locally
- hosting
- sending e-mails
- validate e-mails and passwords
- cleanup goroutine or postgresql service? 
- one connect, pool of connections or what?


