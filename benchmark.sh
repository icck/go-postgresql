#!/bin/bash

echo "=========================================="
echo "PostgreSQL Performance Benchmark"
echo "GORM vs PGX Comparison"
echo "=========================================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Check if PostgreSQL container is already running
if docker compose ps postgres | grep -q "running"; then
    echo "PostgreSQL container is already running."
else
    echo "Starting PostgreSQL container..."
    docker compose up -d
    sleep 5
fi

echo "Waiting for PostgreSQL to be ready..."
until docker compose exec -T postgres pg_isready -U user -d go_database > /dev/null 2>&1; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done

echo "PostgreSQL is ready!"
echo ""

# Run GORM version
echo "=========================================="
echo "Running GORM Version..."
echo "=========================================="
echo ""
cd cmd/gorm
go run main.go
echo ""

# Run PGX version
echo "=========================================="
echo "Running PGX Version..."
echo "=========================================="
echo ""
cd ../pgx
go run main.go
echo ""

echo "=========================================="
echo "Benchmark Complete!"
echo "=========================================="
echo ""
echo "Compare the performance summaries above to see the differences between GORM and PGX."
echo "Key metrics to compare:"
echo "- Seed time (bulk insert performance)"
echo "- Update time (bulk update performance)"
echo "- Delete time (bulk delete performance)"
echo "- Create time (batch create performance)"
echo "- Total time (overall performance)"
