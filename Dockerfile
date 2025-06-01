FROM golang:1.24.3-alpine3.22 AS build
RUN apk update                   \
    && apk upgrade               \
    && apk add --no-cache        \
        zig make ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /app

COPY . .

RUN --mount=type=cache,target=/tmp/.cache/rivabot-build make


FROM scratch AS runner
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/rivabot /
COPY --from=build /app/config.yaml /
ENTRYPOINT ["/rivabot"]

