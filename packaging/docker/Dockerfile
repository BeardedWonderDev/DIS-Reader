FROM debian:12-slim

# Provided by buildx; matches paths GoReleaser places in the temp context (e.g., linux/amd64/dis-agent).
ARG TARGETOS
ARG TARGETARCH

RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
       ca-certificates \
       openjdk-17-jre-headless \
       golang \
    && rm -rf /var/lib/apt/lists/*

# Optional non-root user for parity with prior distroless base.
RUN useradd --create-home --uid 65532 disagent

COPY ${TARGETOS}/${TARGETARCH}/dis-agent /usr/local/bin/dis-agent

USER disagent
ENTRYPOINT ["/usr/local/bin/dis-agent"]
