FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV dbURL = scylladb:9042

CMD go run . -scyllaHost=${dbURL}
EXPOSE 8081