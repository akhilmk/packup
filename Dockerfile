FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the pre-built binary
COPY bin/server /app/server

# Copy the frontend static files
COPY bin/frontend/dist /app/frontend/dist

# Expose port
EXPOSE 8080

# Set environment variable for production
ENV PORT=8080

# Run the binary
CMD ["/app/server"]
