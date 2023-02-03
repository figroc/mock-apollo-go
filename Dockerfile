FROM golang:1.20-alpine as build
WORKDIR /app
COPY . .
RUN go build ./cmd/mock-apollo-go

FROM golang:1.20-alpine
COPY --from=build /app/mock-apollo-go /
ENTRYPOINT ["/mock-apollo-go"]
