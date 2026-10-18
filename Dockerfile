# Build stage
FROM golang:1.22-bookworm AS build

WORKDIR /src

# Copy module files first so Docker caching works
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/dbbackup ./src/cmd/dbbackup

# Final Stage
FROM debian:bookworm-slim

# Install PostgreSQL client tools (pg_dump, pg_restore, pg_isready)
# and CA certificates (for uploading to AWS S3)
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    postgresql-client \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from the build stage
COPY --from=build /out/dbbackup /usr/local/bin/dbbackup

# Set the entrypoint
ENTRYPOINT ["dbbackup"]
CMD ["--help"]
