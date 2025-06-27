FROM ubuntu:22.04
RUN apt update && apt upgrade -y
RUN apt install -y wget curl git build-essential
RUN wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz -O go.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
RUN go version
WORKDIR /app
