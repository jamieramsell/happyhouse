# Build a static Go binary, then run it from a minimal distroless image.
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
# go.sum is copied when present; the skeleton has no external deps yet.
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/auth ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/auth /auth
EXPOSE 8080
ENTRYPOINT ["/auth"]
