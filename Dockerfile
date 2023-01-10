FROM golang:1.19 as go

RUN ["mkdir", "-p", "/test"] 
RUN ["mkdir", "-p", "/test/bin"] 

ADD go.mod go.sum /test
WORKDIR /test
RUN go mod download
ADD e2e.go ./ features ./
# specify each feature folder with go test module, 
# compiles each feature to a binary to be executed, 
# and copies the .feature file with it for runtime.
RUN go test -c -o ./bin/dapp_develop_test ./features/dapp_develop/...
ADD features/dapp_develop/dapp_develop.feature ./bin

FROM stellar/quickstart@sha256:8968d5c3344fe447941ebef12c5bede6f15ba29b63317a488c16c5d5842e4a71

RUN ["mkdir", "-p", "/opt/test"] 
ADD start /opt/test
COPY --from=go /test/bin/ /opt/test/bin
COPY --from=go /usr/local/go/ /usr/local/go/
 
RUN ["mkdir", "-p", "/rust"] 
ENV CARGO_HOME=/rust/.cargo
ENV RUSTUP_HOME=/rust/.rust

ADD install /
RUN ["chmod", "+x", "install"]
RUN /install

ENV PATH="/usr/local/go/bin:$CARGO_HOME/bin:${PATH}"

RUN ["chmod", "+x", "/opt/test/start"]

ENTRYPOINT ["/opt/test/start"]
