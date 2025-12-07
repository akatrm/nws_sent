# Dockerfile for setting up a Solr server with specific version and dependencies
# for now use debian trixie slim as base image
FROM public.ecr.aws/docker/library/debian:trixie-slim
LABEL maintainer="https://github.com/akatrm"

# tied to solr 8.11.2 for compatibility with existing setup
ENV SOLR_VERSION=8.11.2 
ENV SOLR_HOME=/var/solr/data

RUN apt update \
    && apt install -y curl openjdk-21-jre-headless procps

RUN curl -o solr-8.11.4.tgz https://dlcdn.apache.org/lucene/solr/8.11.4/solr-8.11.4.tgz \
    && tar -C /opt -xzf solr-8.11.4.tgz \
    && rm solr-8.11.4.tgz

ENTRYPOINT ["/opt/solr/bin/solr", "start", "-f", "-p", "8983"]