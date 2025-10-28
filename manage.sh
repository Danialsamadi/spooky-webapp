#!/bin/bash

# Webapp Management Script
# Django-style management commands

case "$1" in
    "createsuperuser")
        echo "🚀 Creating superuser..."
        go run manage.go createsuperuser
        ;;
    "cleanusers")
        echo "🗑️ Cleaning users..."
        go run manage.go cleanusers
        ;;
    "listusers")
        echo "📋 Listing users..."
        go run manage.go listusers
        ;;
    "generatecode")
        echo "🎫 Generating invitation code..."
        go run manage.go generatecode
        ;;
    "listcodes")
        echo "🎫 Listing invitation codes..."
        go run manage.go listcodes
        ;;
    "run")
        echo "🚀 Starting web application..."
        go run main.go
        ;;
    "help"|"")
        echo "🚀 Webapp Management CLI"
        echo ""
        echo "Usage: ./manage.sh <command>"
        echo ""
        echo "Available commands:"
        echo "  createsuperuser    - Create a new admin user"
        echo "  cleanusers         - Remove all non-admin users"
        echo "  listusers          - List all users"
        echo "  generatecode       - Generate invitation code"
        echo "  listcodes          - List all invitation codes"
        echo "  run                - Start the web application"
        echo "  help               - Show this help"
        echo ""
        echo "Examples:"
        echo "  ./manage.sh createsuperuser"
        echo "  ./manage.sh cleanusers"
        echo "  ./manage.sh listusers"
        echo "  ./manage.sh generatecode"
        echo "  ./manage.sh listcodes"
        echo "  ./manage.sh run"
        ;;
    *)
        echo "❌ Unknown command: $1"
        echo "Run './manage.sh help' for available commands"
        exit 1
        ;;
esac
