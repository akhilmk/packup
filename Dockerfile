FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the pre-built binary
COPY bin/packup /app/packup

# Copy the frontend static files
COPY bin/frontend/dist /app/frontend/dist

# Copy the migrations folder
COPY bin/migrations /app/migrations

# Expose port
EXPOSE 8080

# Set environment variable for production
ENV PORT=8080

# Run the binary
CMD ["/app/packup"]
