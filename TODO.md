## TODO

* fix the code so that app logs will go to stdout
* start rsyslog, start http sink, and send an event to rsyslog LOCAL1, see if posts while logging its own logs to LOCAL0
* bring down http sink and send an event, see if gets retried
* complete redis saving logic, see if it can save and retrieve metadata
* Dockerize with rsyslog and redis in the same container
* read configs from command line
* make it safely crash when input is wrong
