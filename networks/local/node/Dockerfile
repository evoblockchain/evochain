FROM golang:stretch as build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make git libc-dev bash gcc
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y $PACKAGES

WORKDIR /evochain
# Add source files
COPY . .

# build evochain
RUN make build-linux

# Final image
FROM golang:1.17 as final

WORKDIR /evochaind
# Copy over binaries from the build-env
COPY --from=build-env /evochain/build/evochaind /usr/bin/evochaind
COPY --from=build-env /evochain/build/evochaincli /usr/bin/evochaincli
COPY --from=build-env /evochain/networks/local/node/wrapper.sh /usr/bin/wrapper.sh

EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

