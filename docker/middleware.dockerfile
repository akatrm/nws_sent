# Dockerfile for setting up middleware service
# for now use debian trixie slim as base image
FROM public.ecr.aws/docker/library/debian:trixie-slim
LABEL maintainer="https://github.com/akatrm"

# install compiled go binary
COPY . /middleware

ENTRYPOINT ["middleware"]