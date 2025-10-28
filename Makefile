# Webapp Management Makefile
# Django-style management commands

.PHONY: help createsuperuser cleanusers listusers generatecode listcodes run build

help:
	@echo "🚀 Webapp Management Commands"
	@echo ""
	@echo "Available commands:"
	@echo "  make createsuperuser    - Create a new admin user"
	@echo "  make cleanusers         - Remove all non-admin users"
	@echo "  make listusers          - List all users"
	@echo "  make generatecode       - Generate invitation code"
	@echo "  make listcodes          - List all invitation codes"
	@echo "  make run                - Start the web application"
	@echo "  make build              - Build the application"
	@echo "  make help               - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make createsuperuser"
	@echo "  make cleanusers"
	@echo "  make listusers"
	@echo "  make generatecode"
	@echo "  make listcodes"

createsuperuser:
	@go run manage.go createsuperuser

cleanusers:
	@go run manage.go cleanusers

listusers:
	@go run manage.go listusers

generatecode:
	@go run manage.go generatecode

listcodes:
	@go run manage.go listcodes

run:
	@go run main.go

build:
	@go build -o webapp .
	@echo "✅ Application built successfully!"
	@echo "Run with: ./webapp"
