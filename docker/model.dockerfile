FROM public.ecr.aws/docker/library/debian:trixie-slim
LABEL maintainer="https://github.com/akatrm"

ENV SOLR_VERSION=8.11.2 
ENV SOLR_HOME=/var/solr/data

RUN apt update \
    && apt install -y curl openjdk-21-jre-headless procps

