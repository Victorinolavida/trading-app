FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY backend ./backend
RUN CGO_ENABLED=0 go build -o /out/gateway ./backend/gateway

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/gateway /gateway
EXPOSE 5001
ENTRYPOINT ["/gateway"]
