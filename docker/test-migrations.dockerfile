FROM golang:1.25-alpine

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go test ./pkg/storage/... -v -tags integration -count=1
