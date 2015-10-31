# Poster
Poster is an Rsyslog Omprog module that will take json array from stdout and posts it into user's end point.

http://www.rsyslog.com/doc/v8-stable/configuration/modules/omprog.html

# bring Up http sink
$ sink http

# install
$ sudo mkdir -p /var/spool/poster
$ bin/deploy

# logs
Poster will output logs to STDOUT, in the provided rsyslog config, we instruct rsyslog to put the logs to /tmp/poster.log

# how to run
$ godep go build -o build/poster && echo '[{"user_id":12345,"event":"click","sg_event":"foo"}]' | build/poster
