FROM golang AS base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod verify \
    && go mod download

COPY . .

EXPOSE 4000

FROM base as build

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s" -o ./bin/snippetbox ./cmd/web

FROM base as dev

CMD [ "go", "run", "./cmd/web" ]

FROM alpine AS prod

COPY --from=build /app/bin/snippetbox /app/bin/snippetbox

ENTRYPOINT "/app/bin/snippetbox"  "-dsn" "${SNIPPETBOX_DB_DSN}"

# Metadata
LABEL maintainer="TheAimHero <vsghodekar1@gmail.com>"
LABEL org.opencontainers.image.source="https://github.com/TheAimHero/snippetbox"
