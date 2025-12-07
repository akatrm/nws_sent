# Dockerfile for setting up a Solr server with specific version and dependencies
# for now use debian trixie slim as base image
FROM public.ecr.aws/docker/library/debian:trixie-slim

LABEL maintainer="https://github.com/akatrm"

# install analytics engine wheel package
COPY analytics_engine-*.whl /tmp/


RUN apt update \
    && apt install -y curl python3

