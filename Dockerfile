FROM golang:1.19

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . ./

RUN make build

FROM alpine:3.14
ARG BINPATH="/app/build/scour"
COPY --from=0 $BINPATH /usr/local/bin
