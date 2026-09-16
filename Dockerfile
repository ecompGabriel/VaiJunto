FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .

RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/driver-client ./cmd/driver-client
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/passenger-client ./cmd/passenger-client

FROM alpine:3.22

WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/driver-client /app/driver-client
COPY --from=build /out/passenger-client /app/passenger-client

EXPOSE 8080
CMD ["/app/server"]
