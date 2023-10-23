# -----------------------------------------------------------------------------
# Stages
# -----------------------------------------------------------------------------

ARG IMAGE_GO_BUILDER=golang:1.21.0-bullseye@sha256:02f350d8452d3f9693a450586659ecdc6e40e9be8f8dfc6d402300d87223fdfa
ARG IMAGE_FINAL=senzing/senzingapi-runtime:3.7.1

# -----------------------------------------------------------------------------
# Stage: go_builder
# -----------------------------------------------------------------------------

FROM ${IMAGE_GO_BUILDER} as go_builder
ENV REFRESHED_AT=2023-10-02
LABEL Name="senzing/validate-builder" \
      Maintainer="support@senzing.com" \
      Version="0.0.4"

# Copy local files from the Git repository.

COPY ./rootfs /
COPY . ${GOPATH}/src/validate

# Set path to Senzing libs.

ENV LD_LIBRARY_PATH=/opt/senzing/g2/lib/

# Build go program.

WORKDIR ${GOPATH}/src/validate
RUN make build

# Copy binaries to /output.

RUN mkdir -p /output \
 && cp -R ${GOPATH}/src/validate/target/*  /output/

# -----------------------------------------------------------------------------
# Stage: final
# -----------------------------------------------------------------------------

FROM ${IMAGE_FINAL} as final
ENV REFRESHED_AT=2023-10-03
LABEL Name="senzing/validate" \
      Maintainer="support@senzing.com" \
      Version="0.0.4"

# Copy files from prior stage.

COPY --from=go_builder "/output/linux-amd64/validate" "/app/validate"

# Runtime environment variables.

ENV LD_LIBRARY_PATH=/opt/senzing/g2/lib/

# Runtime execution.

WORKDIR /app
ENTRYPOINT ["/app/validate"]
