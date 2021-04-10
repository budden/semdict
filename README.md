# Semantic dictionary - a sense-based English-Russian dictionary

# Quick start

```bash
# run once at the start of work
make setup
# run every time migrations changes?
make up
# run server
make run
```

## Concept

### Idea of "semantic dictionary"

There are two differences to normal Language A to Language B translation dictionary:

- there is a cell per sense of the word, not a cell per word. Words having multiple meanings have multiple senses.
- there are many variants of translation, and we track the sources of translation. E.g. Oracle and Microsoft can use different translations of some word sense to Russian. We create "Oracle" dialect to store Oracle's translations and "Microsoft" dialect to store Microsoft's.

### Detailed requirement specifications (in Russian)

[Look here](doc/тз/общее.md)

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
