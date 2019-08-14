# Semantic dictionary - an aristocratic communism multilingual glossary engine  

## Concept

### Golang exercise
The main goal of this project is to train myself in a full-stack development using golang and postgresql. 
You know golang is usually associated with «microservices», but I consider golang as just a good high-level programming language. So here I use golang as a replacement for a PHP.

### Idea of "semantic dictionary"
Dictionaries are used to translate words from one language to another. But every word usually have 
multiple meanings (senses), so word translations are NxN relationships between words. We're building a translation
engine where the key is the pair of the word and the defining phrase which disambiguates the sense. 
This way, there is a chance that glossary entry to glossary entry translation is unambiguous. 

### Why communism?
We aim for the minimization of the administrative burden. Git is an example 
of excellent approach to that and we tried to implement git-like responsibility
structure, where a database of senses for each language or dialect is 
owned by only one person (like git repo), and any other person can fork 
the language and suggest his/her changes. But we forgot about translations
:) Who is responsible for English-Russian translation? This way we found
that there is no clean responsibility bounds and decided to reject
the entire ownership concept. So, imagine all the people sharing all the world.
All registered people, of course. 

### Why aristocratic
All animals are equal but others are more equal. Any of moderators can undo the change history to some point in the past. Maybe in the future we will be able to introduce more sophisticated moderation, but not now, because we are limited by time severely. 

### Detailed requirement specifications (outdated)

[Look here](https://bitbucket.org/budden/ppr/src/master/док/словарь.md?at=master&fileviewer=file-view-default)

## State
Pre-alpha, no deployment. Not all features are implemented.

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
