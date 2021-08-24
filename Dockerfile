FROM golang:1.16 AS build-stage
ADD ./ /graphqld
RUN cd /graphqld && go build ./cmd/graphqld

FROM ubuntu
ARG DEBIAN_FRONTEND=noninteractive
RUN apt update \
    && apt install --no-install-suggests --no-install-recommends -y unzip \ 
    curl wget iputils-ping python3 nodejs \
    && apt-get upgrade -y -o Dpkg::Options::="--force-confold" \
    && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
RUN mkdir -p /var/graphqld

COPY --from=build-stage /graphqld/graphqld /bin/graphqld

VOLUME ["/var/graphqld"]
EXPOSE 80

ENTRYPOINT ["graphqld"]