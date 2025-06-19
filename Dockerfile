# syntax=docker/dockerfile:1-labs
FROM --platform=$BUILDPLATFORM node:22 AS web-builder
WORKDIR /app/web
COPY web .
RUN make build
FROM --platform=$BUILDPLATFORM golang:1.24 AS go-builder
WORKDIR /app
COPY --exclude=web . .
COPY --from=web-builder /app/web/dist web/dist
RUN make build-go@linux
FROM alpine
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
COPY --from=go-builder /app/out/$TARGETOS-$TARGETARCH/* .
ENV DRPP_IS_IN_CONTAINER=true
ENTRYPOINT ["./drpp"]
EXPOSE 8040
