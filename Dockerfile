# syntax=docker/dockerfile:1.7

FROM golang:1.26.4-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/web ./cmd/web \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/worker ./cmd/worker \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/restore-season ./cmd/restore-season \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/provider-stubs ./cmd/provider-stubs

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/web /app/web
COPY --from=build /out/worker /app/worker
COPY --from=build /out/migrate /app/migrate
COPY --from=build /out/restore-season /app/restore-season
COPY --from=build /out/provider-stubs /app/provider-stubs
COPY migrations /app/migrations
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/web"]
