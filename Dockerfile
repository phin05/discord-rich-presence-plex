FROM --platform=$BUILDPLATFORM node:24 AS web-builder
WORKDIR /app
COPY . .
ENV CONTAINERISED_BUILD=true
RUN make install-deps-web build-web

FROM --platform=$BUILDPLATFORM golang:1.26 AS server-builder
WORKDIR /app
COPY . .
COPY --from=web-builder /app/web/dist web/dist
RUN make build-server@linux

FROM alpine
WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
COPY --from=server-builder /app/out/$TARGETOS-$TARGETARCH/* .
ENV DRPP_CONTAINERISED=true
ENTRYPOINT ["./drpp"]
EXPOSE 8040
