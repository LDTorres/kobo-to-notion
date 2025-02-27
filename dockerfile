FROM golang:1.24.0

ENV DEBIAN_FRONTEND=noninteractive

RUN apt clean && \
    rm -rf /var/lib/apt/lists/* && \
    apt update --fix-missing && \
    apt install -y apt-utils ca-certificates && \
    update-ca-certificates

RUN apt update
RUN apt install -y curl gcc-arm-linux-gnueabi --fix-missing

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["/bin/bash"]