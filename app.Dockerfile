FROM ubuntu-golang
WORKDIR /app
COPY . .
VOLUME [ "/data" ]
CMD [ "go", "run", "server/main.go" ]