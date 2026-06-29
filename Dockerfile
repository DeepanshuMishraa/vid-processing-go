#Build 
FROM golang:1.26 AS builder 

WORKDIR /app 

COPY go.sum go.mod ./ 

RUN go mod download 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go


#RUN

FROM alpine:latest 

WORKDIR /app 

COPY --from=builder /app/main . 
COPY --from=builder /app/migrations ./migrations

EXPOSE 3001

ENTRYPOINT ["./main"]
