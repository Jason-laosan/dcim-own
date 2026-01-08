#!/bin/bash

# OPC Alert System Quick Start Script

echo "====================================="
echo "OPC Alert System - Quick Start"
echo "====================================="
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Start infrastructure services
echo "Step 1: Starting infrastructure services (PostgreSQL, InfluxDB, Kafka)..."
docker-compose up -d

echo "Waiting for services to be ready..."
sleep 15

# Check if services are healthy
echo ""
echo "Step 2: Checking service status..."
docker-compose ps

echo ""
echo "Step 3: Building Maven projects..."
mvn clean package -DskipTests

if [ $? -ne 0 ]; then
    echo "Error: Maven build failed"
    exit 1
fi

echo ""
echo "====================================="
echo "Setup Complete!"
echo "====================================="
echo ""
echo "Infrastructure Services:"
echo "  - PostgreSQL: localhost:5432"
echo "  - InfluxDB: http://localhost:8086"
echo "  - Kafka: localhost:9092"
echo "  - Kafka UI: http://localhost:8080"
echo ""
echo "InfluxDB Credentials:"
echo "  - Username: admin"
echo "  - Password: admin123456"
echo "  - Token: my-super-secret-auth-token"
echo "  - Organization: opc_organization"
echo "  - Bucket: opc_data"
echo ""
echo "PostgreSQL Credentials:"
echo "  - Database: opc_alert_db"
echo "  - Username: postgres"
echo "  - Password: postgres123"
echo ""
echo "Next Steps:"
echo "  1. Start Alert Consumer:"
echo "     cd alert-consumer && mvn spring-boot:run"
echo ""
echo "  2. Start Alert Engine (in another terminal):"
echo "     cd alert-engine && mvn spring-boot:run"
echo ""
echo "  3. Configure alert rules in PostgreSQL database"
echo ""
echo "To stop all services: docker-compose down"
echo "====================================="
