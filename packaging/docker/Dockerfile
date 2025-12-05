FROM gcr.io/distroless/base-debian12:nonroot

# Provided by buildx; matches paths GoReleaser places in the temp context (e.g., linux/amd64/dis-agent).
ARG TARGETOS
ARG TARGETARCH

COPY ${TARGETOS}/${TARGETARCH}/dis-agent /usr/local/bin/dis-agent

USER nonroot
ENTRYPOINT ["/usr/local/bin/dis-agent"]
