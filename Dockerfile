FROM alpine:3.6 as helmbuild

RUN apk add --update --no-cache \
ca-certificates \
curl \
git \
gzip \
tar
ARG VERSION=v2.9.1
ARG FILENAME=helm-${VERSION}-linux-amd64.tar.gz
WORKDIR /
RUN curl -L "https://storage.googleapis.com/kubernetes-helm/${FILENAME}" | tar zxv -C /tmp

FROM golang:1.11 as gobuild
COPY --from=helmbuild /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /boondoggle

# The image we keep
FROM google/cloud-sdk:alpine
RUN gcloud components install kubectl
RUN apk add --update --no-cache git bash curl
COPY --from=helmbuild /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=helmbuild /tmp/linux-amd64/helm /bin/helm
COPY --from=gobuild /boondoggle /bin/boondoggle
RUN helm init -c
RUN helm plugin install https://github.com/futuresimple/helm-secrets
CMD ["boondoggle", "-h"]
