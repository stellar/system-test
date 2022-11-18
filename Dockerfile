FROM golang:1.19 as go

RUN ["mkdir", "-p", "/test"] 
RUN ["mkdir", "-p", "/test/bin"] 
ADD features /test/features
ADD go.mod /test
ADD go.sum /test
ADD e2e.go /test

WORKDIR /test
# specify each feature folder with go test module, 
# compiles each feature to a binary to be executed, 
# and copies the .feature file with it for runtime.
RUN go test -c -o ./bin/dapp_develop_test ./features/dapp_develop/...
ADD features/dapp_develop/dapp_develop.feature ./bin

FROM stellar/quickstart:soroban-dev

RUN ["mkdir", "-p", "/opt/test"] 
ADD start /opt/test
COPY --from=go /test/bin/ /opt/test/bin

RUN ["mkdir", "-p", "/rust"] 
ENV CARGO_HOME=/rust/.cargo
ENV RUSTUP_HOME=/rust/.rust

ADD install /
RUN ["chmod", "+x", "install"]
RUN /install

ENV PATH="$CARGO_HOME/bin:${PATH}"

RUN ["chmod", "+x", "/opt/test/start"]

ENTRYPOINT ["/opt/test/start"]
