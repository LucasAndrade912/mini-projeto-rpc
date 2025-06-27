FROM ubuntu-golang
WORKDIR /app
COPY pkg ./pkg
COPY client ./client
COPY go.mod .