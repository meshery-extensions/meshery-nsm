FROM golang:1.23-alpine AS builder

ARG VERSION
ARG GIT_COMMITSHA
WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download && go mod verify

# Copy the go sourcee
COPY main.go main.go
COPY internal/ internal/
COPY nsm/ nsm/

RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build \
    -ldflags="-w -s -X main.version=$VERSION -X main.gitsha=$GIT_COMMITSHA" \
    -trimpath \
    -a \
    -o meshery-nsm main.go


FROM gcr.io/distroless/static:nonroot
WORKDIR /home/nonroot/.meshery

COPY --from=builder --chown=65532:65532 /build/meshery-nsm .


ENV DISTRO="debian"

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/meshery-nsm", "-version"]

ENTRYPOINT ["./meshery-nsm"]
