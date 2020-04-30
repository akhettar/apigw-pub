## Base build image
FROM golang:alpine3.11 AS build_base

# Install some dependencies needed to build the project
USER root
RUN apk add bash ca-certificates git gcc g++ libc-dev
WORKDIR /go/src/apigw-pub

# Force the go compiler to use modules
ENV GO111MODULE=on

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

#This is the ‘magic’ step that will download all the dependencies that are specified in
RUN go mod download

# This image builds the swagger publisher tool
FROM build_base AS tool_builder
# Here we copy the rest of the source code
COPY . .
COPY docker-entrypoint.sh /go/bin/

# And compile the project
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go install -a -tags netgo -ldflags '-w -extldflags "-static"' .

#In this last stage, we start from a fresh Alpine image, to reduce the image size and not ship the Go compiler in our production artifacts.
FROM alpine AS swaggerPublisher

# Finally we copy the statically compiled Go binary and the docker entry point
COPY --from=tool_builder /go/bin/apigw-pub /bin/apigw-pub
COPY --from=tool_builder /go/bin/docker-entrypoint.sh /bin/docker-entrypoint.sh
