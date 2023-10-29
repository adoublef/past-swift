# syntax=docker/dockerfile:1

ARG GO_VERSION=1.21

FROM golang:${GO_VERSION} AS base

WORKDIR /usr/src

COPY go.* .
RUN go mod download

COPY . .

FROM base AS test

RUN go test -v -cover -count 1 ./...

FROM base AS build

# build application
# cgo needed for litefs
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -extldflags '-static'" \
    -buildvcs=false \
    -tags osusergo,netgo \
    -o /usr/bin/a ./cmd/past-swift/

# build migration
# cgo needed for litefs
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -extldflags '-static'" \
    -buildvcs=false \
    -tags osusergo,netgo \
    -o /usr/bin/b ./cmd/migrations/

FROM alpine AS deploy

WORKDIR /opt

ARG LITEFS_CONFIG="litefs.yml"
ENV LITEFS_DIR="/litefs"
ENV DATABASE_URL="${LITEFS_DIR}/database.db"
ENV DATABASE_URL_SESSIONS="${LITEFS_DIR}/sessions.db"
ENV INTERNAL_PORT=8080
ENV PORT=8081

# copy binary from build
COPY --from=build /usr/bin/a /usr/bin/b ./

# install sqlite, ca-certificates, curl and fuse for litefs
RUN apk add --no-cache bash fuse3 sqlite ca-certificates curl

# prepar for litefs
COPY --from=flyio/litefs:0.5 /usr/local/bin/litefs /usr/local/bin/litefs
ADD litefs/${LITEFS_CONFIG} /etc/litefs.yml
RUN mkdir -p /data ${LITEFS_DIR}

FROM deploy AS local

# prepare for infisical
# RUN curl -1sLf \
#     'https://dl.cloudsmith.io/public/infisical/infisical-cli/setup.alpine.sh' | bash \
#     && apk add infisical

ENTRYPOINT ["litefs", "mount"]

FROM deploy AS final

ENTRYPOINT ["litefs", "mount"]