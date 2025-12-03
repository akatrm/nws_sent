FROM public.ecr.aws/docker/library/debian:trixie-slim
LABEL maintainer="https://github.com/akatrm"

ENV SOLR_VERSION=1.21
ENV SOLR_HOME=/var/solr/data

RUN apt update \
    && apt install -y curl openjdk-21-jre-headless procps

RUN curl -o apache-nutch-1.21-bin.tar.gz https://dlcdn.apache.org/nutch/1.21/apache-nutch-1.21-bin.tar.gz \
    && tar -C /opt -xzf apache-nutch-1.21-bin.tar.gz  \
    && rm apache-nutch-1.21-bin.tar.gz

ENTRYPOINT ["/opt/nutch/bin/solr", "start", "-f", "-p", "8983"]