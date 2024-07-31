# stage for development, which contains tools for code generation and debugging.
FROM golang:1.22.2-bullseye as builder

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    wget \
    make \
    unzip \
    git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# install protoc
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.20.1/protoc-3.20.1-linux-x86_64.zip \
    && unzip -d /usr/local protoc-3.20.1-linux-x86_64.zip \
    && rm protoc-3.20.1-linux-x86_64.zip

# install protoc plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.10.3
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.10.3

WORKDIR /work

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . ./

RUN go build -o /openve ./go/cmd



# runner stage for server
FROM debian:bullseye-slim as runner

COPY --from=builder /openve /openve

CMD ["/openve"]
