# mysql-tcp-analyzer

tools to analyze MySQL performance via tcpdump


## usage

This expects a pcap preprocessed with `tshark` piped into stdin:

```
# collect tcpdump from mysql server
sudo tcpdump -i any -G 15 -W 1 -w mysql.pcap 'port 3306'

# preprocess tcpdump
tshark -r mysql.pcap \
  -Y mysql -Tjson \
  -e tcp.flags.fin \
  -e tcp.flags.reset \
  -e tcp.analysis.lost_segment \
  -e tcp.analysis.ack_lost_segment \
  -e frame.number \
  -e frame.time_relative \
  -e tcp.stream \
  -e mysql.command \
  -e mysql.query \
  -e mysql.payload \
  -e mysql.response_code > mysql-tcp.json

  # run tool (in normalized-transactions mode)
  make && bin/analyze --mode normalized-transactions < mysql-tcp.json > normalized-transactions.json
```