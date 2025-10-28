#!/bin/bash

# Production Server Management Script
# For managing the webapp on your server

APP_NAME="webapp"
APP_DIR="/path/to/your/webapp"  # Change this to your actual path
LOG_FILE="/var/log/webapp.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to log messages
log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a $LOG_FILE
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a $LOG_FILE
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a $LOG_FILE
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a $LOG_FILE
}

# Function to check if app is running
is_running() {
    pgrep -f "$APP_NAME" > /dev/null
}

# Function to start the application
start_app() {
    if is_running; then
        warning "Application is already running"
        return 1
    fi
    
    log "Starting $APP_NAME..."
    cd $APP_DIR
    nohup ./$APP_NAME > $LOG_FILE 2>&1 &
    sleep 2
    
    if is_running; then
        log "Application started successfully (PID: $(pgrep -f $APP_NAME))"
    else
        error "Failed to start application"
        return 1
    fi
}

# Function to stop the application
stop_app() {
    if ! is_running; then
        warning "Application is not running"
        return 1
    fi
    
    log "Stopping $APP_NAME..."
    pkill -f "$APP_NAME"
    sleep 2
    
    if ! is_running; then
        log "Application stopped successfully"
    else
        error "Failed to stop application"
        return 1
    fi
}

# Function to restart the application
restart_app() {
    log "Restarting $APP_NAME..."
    stop_app
    sleep 1
    start_app
}

# Function to show status
show_status() {
    if is_running; then
        log "Application is running (PID: $(pgrep -f $APP_NAME))"
    else
        warning "Application is not running"
    fi
}

# Function to show logs
show_logs() {
    if [ -f "$LOG_FILE" ]; then
        tail -f $LOG_FILE
    else
        error "Log file not found: $LOG_FILE"
    fi
}

# Function to create admin user
create_admin() {
    cd $APP_DIR
    log "Creating admin user..."
    go run manage.go createsuperuser
}

# Function to clean users
clean_users() {
    cd $APP_DIR
    warning "This will delete all non-admin users!"
    read -p "Are you sure? (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
        log "Cleaning users..."
        go run manage.go cleanusers
    else
        info "Operation cancelled"
    fi
}

# Function to list users
list_users() {
    cd $APP_DIR
    go run manage.go listusers
}

# Function to generate invite code
generate_code() {
    cd $APP_DIR
    go run manage.go generatecode
}

# Function to list invite codes
list_codes() {
    cd $APP_DIR
    go run manage.go listcodes
}

# Function to build the application
build_app() {
    cd $APP_DIR
    log "Building application..."
    go build -o $APP_NAME .
    if [ $? -eq 0 ]; then
        log "Build successful"
    else
        error "Build failed"
        return 1
    fi
}

# Function to update and restart
update_app() {
    cd $APP_DIR
    log "Updating application..."
    
    # Pull latest code (if using git)
    # git pull origin main
    
    # Build new version
    build_app
    
    # Restart application
    restart_app
}

# Main script logic
case "$1" in
    "start")
        start_app
        ;;
    "stop")
        stop_app
        ;;
    "restart")
        restart_app
        ;;
    "status")
        show_status
        ;;
    "logs")
        show_logs
        ;;
    "build")
        build_app
        ;;
    "update")
        update_app
        ;;
    "createsuperuser")
        create_admin
        ;;
    "cleanusers")
        clean_users
        ;;
    "listusers")
        list_users
        ;;
    "generatecode")
        generate_code
        ;;
    "listcodes")
        list_codes
        ;;
    "help"|"")
        echo -e "${BLUE}🚀 Webapp Server Management${NC}"
        echo ""
        echo "Usage: $0 <command>"
        echo ""
        echo -e "${GREEN}Application Management:${NC}"
        echo "  start              - Start the application"
        echo "  stop               - Stop the application"
        echo "  restart            - Restart the application"
        echo "  status             - Show application status"
        echo "  logs               - Show application logs"
        echo "  build              - Build the application"
        echo "  update             - Update and restart application"
        echo ""
        echo -e "${GREEN}User Management:${NC}"
        echo "  createsuperuser    - Create a new admin user"
        echo "  cleanusers         - Remove all non-admin users"
        echo "  listusers          - List all users"
        echo "  generatecode       - Generate invitation code"
        echo "  listcodes          - List all invitation codes"
        echo ""
        echo -e "${GREEN}Examples:${NC}"
        echo "  $0 start"
        echo "  $0 createsuperuser"
        echo "  $0 restart"
        echo "  $0 logs"
        ;;
    *)
        error "Unknown command: $1"
        echo "Run '$0 help' for available commands"
        exit 1
        ;;
esac
