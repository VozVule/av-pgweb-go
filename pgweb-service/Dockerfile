ARG GO_VERSION=1.22.2
ARG APP_CMD=./cmd/api
ARG INSTALL_ATLAS=false

# --- Build stage ---
FROM golang:${GO_VERSION}-alpine AS builder
ARG APP_CMD

WORKDIR /src

# Copy module files first to leverage Docker layer caching for dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the remaining source.
COPY . .

# Build a static binary so we can use a tiny runtime image.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /bin/pgweb ${APP_CMD}

# --- Runtime stage ---
FROM alpine:3.20
ARG INSTALL_ATLAS

WORKDIR /app
RUN adduser -D -g '' pgweb \
	&& apk add --no-cache ca-certificates \
	&& if [ "$INSTALL_ATLAS" = "true" ]; then \
		apk add --no-cache curl; \
		curl -sSfL https://release.ariga.io/atlas/atlas-linux-amd64-latest -o /usr/local/bin/atlas; \
		chmod +x /usr/local/bin/atlas; \
		apk del curl; \
	fi

COPY --from=builder /bin/pgweb /app/pgweb
COPY --from=builder /src/migrations /app/migrations

USER pgweb
EXPOSE 8080
ENTRYPOINT ["/app/pgweb"]
