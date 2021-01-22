FROM golang:1.15 as builder

WORKDIR /ratelimit-sample
COPY . /ratelimit-sample

ENV CGO_ENABLED=0
ENV GO111MODULE=on
RUN go build -mod=readonly -v -o server main.go

FROM alpine:3

RUN apk add --no-cache ca-certificates
RUN mkdir -p /ratelimit-sample
WORKDIR /ratelimit-sample

COPY --from=builder /ratelimit-sample/server /ratelimit-sample/server

# Create appuser.
ENV USER=appuser
ENV UID=51197

RUN adduser \
--disabled-password \
--gecos "application user" \
--no-create-home \
--uid "${UID}" \
"${USER}"

RUN chown appuser:appuser /ratelimit-sample
RUN chown appuser:appuser /ratelimit-sample/*
USER appuser:appuser

ENTRYPOINT ["/ratelimit-sample/server"]