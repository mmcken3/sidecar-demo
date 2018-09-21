FROM golang:1.10

ARG package

WORKDIR /go/src/github.com/mmcken3/sidecar-demo/cmd
COPY . /go/src/github.com/mmcken3/sidecar-demo/cmd
RUN go install "github.com/mmcken3/sidecar-demo/$package"