# -----------------------------------------------------------------------------
# Stages
# -----------------------------------------------------------------------------

ARG IMAGE_BUILDER=golang:1.25.1-bookworm@sha256:c423747fbd96fd8f0b1102d947f51f9b266060217478e5f9bf86f145969562ee
ARG IMAGE_FINAL=senzing/senzingsdk-runtime:4.0.0@sha256:332d2ff9f00091a6d57b5b469cc60fd7dc9d0265e83d0e8c9e5296541d32a4aa

# -----------------------------------------------------------------------------
# Stage: builder
# -----------------------------------------------------------------------------

FROM ${IMAGE_BUILDER} AS builder
ENV REFRESHED_AT=2024-07-01
LABEL Name="senzing/go-builder" \
      Maintainer="support@senzing.com" \
      Version="0.1.0"

# Run as "root" for system installation.

USER root

# Copy local files from the Git repository.

COPY ./rootfs /
COPY . ${GOPATH}/src/validate

# Set path to Senzing libs.

ENV LD_LIBRARY_PATH=/opt/senzing/er/lib/

# Build go program.

WORKDIR ${GOPATH}/src/validate
RUN make build

# Copy binaries to /output.

RUN mkdir -p /output \
 && cp -R ${GOPATH}/src/validate/target/*  /output/

# -----------------------------------------------------------------------------
# Stage: final
# -----------------------------------------------------------------------------

FROM ${IMAGE_FINAL} AS final
ENV REFRESHED_AT=2024-07-01
LABEL Name="senzing/template-go" \
      Maintainer="support@senzing.com" \
      Version="0.0.1"
HEALTHCHECK CMD ["/app/healthcheck.sh"]
USER root

# Install packages via apt-get.

# Copy files from repository.

COPY ./rootfs /

# Copy files from prior stage.

COPY --from=builder /output/linux/validate /app/validate

# Run as non-root container

USER 1001

# Runtime environment variables.

ENV LD_LIBRARY_PATH=/opt/senzing/er/lib/

# Runtime execution.

WORKDIR /app
ENTRYPOINT ["/app/validate"]
