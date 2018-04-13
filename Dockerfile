
FROM alpine:3.6 as build

RUN apk add --update --no-cache \
ca-certificates \ 
curl \
git \
gzip \
tar

ARG VERSION=v2.8.1
ARG FILENAME=helm-${VERSION}-linux-amd64.tar.gz

WORKDIR /

RUN curl -L "https://storage.googleapis.com/kubernetes-helm/${FILENAME}" | tar zxv -C /tmp

# The image we keep
FROM google/cloud-sdk:alpine
RUN gcloud components install kubectl
RUN apk add --update --no-cache git ca-certificates bash curl

COPY --from=build /tmp/linux-amd64/helm /bin/helm
COPY boondoggle /bin/boondoggle

CMD ["boondoggle", "-h"]
