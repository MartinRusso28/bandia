FROM golang:1.27.1-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -o /bandia ./cmd/bandia

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /bandia /app/bandia
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/bandia"]
CMD ["serve"]
