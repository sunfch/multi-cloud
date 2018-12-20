FROM ubuntu:16.04
RUN  apt-get update
RUN  apt-get install python2.7 -y

COPY datamover /datamover
ENTRYPOINT ["/datamover"]
