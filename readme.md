# Mikrotik Guard

Receives remote log messages, if the message contents text 
defined in ```phrases``` array it will be sent trough Telegram to all
authorized users.

## Environment variables:
* ```TG_TOKEN``` - Telegram bot token (use @BotFather to get it)
* ```TG_PASSWORD``` - Say this word to authorize user
* ```LOGGER_BIND``` - address and port for bind Syslog server, ex: ```0.0.0.0:514```
* ```DATA_DIR``` 	- folder, that contains users.json file, RW permissions needed

Search phrases are defined in the ```phrases.go``` file, feel free to change it

## Mikrotik settings:
```
/system logging action
add bsd-syslog=yes name=guard remote=GUARD_ADDRESS \
    remote-port=GUARD_PORT target=remote
/system logging
add action=guard topics=system
add action=guard topics=account
add action=guard topics=critical

```