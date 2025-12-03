FROM public.ecr.aws/docker/library/debian:trixie-slim
LABEL maintainer="https://github.com/akatrm"


COPY . /middleware

ENTRYPOINT ["middleware"]