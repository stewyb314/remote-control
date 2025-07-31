FROM alpine:latest

RUN apk update && \
    apk upgrade && \
    apk add --no-cache mariadb-client bash git make musl-dev go openrc

RUN mkdir -p /remote-control && \
    mkdir -p /src/internal && \
    mkdir -p /src/protos && \
    mkdir -p /src/cmd

WORKDIR /src
COPY internal ./internal
COPY ../cmd ./cmd
COPY ../protos ./protos
COPY ../Makefile ./
COPY ../go.mod ./
RUN go mod tidy && make agent && mv agent /remote-control/agent

WORKDIR /remote-control
RUN mkdir -p /src/test-scripts
COPY ../test-scripts ./test-scripts
RUN chmod a+x ./test-scripts/*.sh
ENTRYPOINT [ "./agent" ]