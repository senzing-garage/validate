# -----------------------------------------------------------------------------
# Stages
# -----------------------------------------------------------------------------

ARG IMAGE_BUILDER=golang:1.25.4-bookworm@sha256:a4e11d718c8aadd9aa8382d328dcb9cbf73b9b4a4f508f47f821ee429fd687cf
ARG IMAGE_FINAL=senzing/senzingsdk-runtime:4.1.0@sha256:e57d751dc0148bb8eeafedb7accf988413f50b54a7e46f25dfe4559d240063e5

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
