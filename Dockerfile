FROM ubuntu:14.04

ADD azkvbs /
ENTRYPOINT ["/azkvbs"]
