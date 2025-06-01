FROM golang:1.24.3-alpine3.22 AS build
RUN apk update                  \
    && apk upgrade              \
    && apk add --no-cache       \
        zig make                \
    && rm -rf /var/cache/apk/*

WORKDIR /app

COPY . .

RUN --mount=type=cache,target=/tmp/.cache/rivabot-build make


FROM scratch AS bot
COPY --from=build /app/rivabot /
COPY --from=build /app/config.yaml /
CMD ["/rivabot"]

