FROM golang:1.22

WORKDIR /build

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go install github.com/magefile/mage@v1.15.0
RUN mage go:build

FROM scratch
COPY --from=0 /build/target/bin/devnet-explorer /devnet-explorer

CMD [ "/devnet-explorer" ]
