# Semantic dictionary - an anarchic multilingual glossary engine  

## Concept

### Golang exercise
The main goal of this project is to train myself in a full-stack development using golang and postgresql. 
You know golang is usually associated with «microservices», but I consider golang as just a good 
high-level programming language. So here I use golang as a replacement for a PHP.

### Idea of "semantic dictionary"
Dictionaries are used to translate words from one language to another. But every word usually have 
multiple meanings (senses), so word translations are NxN relationships between words. We're building a translation
engine where the key is the pair of the word and the defining phrase which disambiguates the sense. 
This way, there is a chance that glossary entry to glossary entry translation is unambiguous. 

### Anarchic = like github
Github has a "fork" feature. That means that anyone can create one's own version of every project. Operations on forks
are "compare versions" and "request an original author to accept my changes", or «Pull request».
We plan to support two sorts of forks:

- make a new dialect; for instance, localized MS Windows can have different translations for the same sense compared to Linux or Android.
So we create three forks on "Russian" language and call them "Windows localization", "Linux localization" and "Android localization"
- do a collective work; for now, fork (or branch) with a subsequent pull request seem to be a good way to administer collective work.
Time will tell if it is really so good.

## State
Pre-alpha. You can register on the [semantic-dict.ru](semantic-dict.ru) already, but at any time I can 
deploy a new database with a different structure and your registration is gone. 

## Technology

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
