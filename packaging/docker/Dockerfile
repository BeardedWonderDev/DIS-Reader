FROM gcr.io/distroless/base-debian12:nonroot

ARG TARGETARCH
COPY dis-agent /usr/local/bin/dis-agent

USER nonroot
ENTRYPOINT ["/usr/local/bin/dis-agent"]
