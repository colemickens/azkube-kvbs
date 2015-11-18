FROM ubuntu:14.04

# RUN bash -c "\
# 	sudo apt-get update \
# 	&& sudo apt-get upgrade \
# 	&& sudo apt-get install openssl \
# 	&& apt-get clean \
# 	&& rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*"

ADD azkvbs /

CMD ["/azkvbs"]
