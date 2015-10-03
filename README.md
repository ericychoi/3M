# Poster
It posts events.

# config for rsyslog
module(load="omprog")
        local0.* action(type="omprog" binary="/opt/sendgrid/prism/current/prism -config_path=/etc/sendgrid/prism.properties" output="/var/log/sendgrid/prism.log" hup.signal="TERM" queue.spoolDirectory="/var/spool/rsyslog" queue.filename="prism_kafka_queue" queue.type="Disk" queue.maxdiskspace="1G")

http://www.rsyslog.com/doc/v8-stable/configuration/modules/omprog.html

# Bring Up http sink
$ sink http
