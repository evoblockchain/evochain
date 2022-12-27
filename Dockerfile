# Simple usage with a mounted data directory:
# > docker build -t evochain .
# > docker run -it -p 36657:36657 -p 36656:36656 -v ~/.evochaind:/root/.evochaind -v ~/.evochaincli:/root/.evochaincli evochain evochaind init mynode
# > docker run -it -p 36657:36657 -p 36656:36656 -v ~/.evochaind:/root/.evochaind -v ~/.evochaincli:/root/.evochaincli evochain evochaind start
FROM golang:1.17.2-alpine AS build-env

# Install minimum necessary dependencies, remove packages
RUN apk add --no-cache curl make git libc-dev bash gcc linux-headers eudev-dev

# Set working directory for the build
WORKDIR /go/src/github.com/evoblockchain/evochain

# Add source files
COPY . .

ENV GO111MODULE=on \
    GOPROXY=http://goproxy.cn
# Build EVOChain
RUN make install

# Final image
FROM alpine:edge

WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/evochaind /usr/bin/evochaind
COPY --from=build-env /go/bin/evochaincli /usr/bin/evochaincli

# Run evochaind by default, omit entrypoint to ease using container with evochaincli
CMD ["evochaind"]
