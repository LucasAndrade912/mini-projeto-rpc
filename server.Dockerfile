FROM ubuntu-golang
WORKDIR /app
COPY pkg ./pkg
COPY server ./server
COPY go.mod .
COPY logs.txt .
COPY snapshots ./snapshots
VOLUME [ "/data" ]
EXPOSE 5000
CMD [ "go", "run", "server/main.go" ]