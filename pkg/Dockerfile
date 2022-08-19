FROM --platform=$BUILDPLATFORM golang:1.18-alpine as builder
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ADD . /src
WORKDIR /src
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags "-X main.Version=$VERSION" -a -o movie-data-api .
FROM alpine
COPY --from=builder /src /app
COPY --from=builder /src/movie-data-api /app/movie-data-api
WORKDIR /app
EXPOSE 8080
ENTRYPOINT ["/app/movie-data-api"]

