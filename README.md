# Haunted Blog - Go Web Application

A spooky blog application built with Go, featuring ghost mode theme, authentication, and comprehensive logging.

## Features

- **Ghost Mode Theme** - Dark, spooky interface with glowing effects
- **Ghost Animations** - Floating ghosts and spooky loading screens
- **User Authentication** - Login, signup, logout with session management
- **CRUD Operations** - Create, read, update, delete blog posts
- **Comprehensive Logging** - Track all user activities and system events
- **Docker Support** - Fully containerized application

## Quick Start with Docker

### Prerequisites
- Docker
- Docker Compose

### 1. Clone and Navigate
```bash
cd /path/to/webapp
```

### 2. Build and Run
```bash
# Build and start all services
docker-compose up --build

# Run in background
docker-compose up -d --build
```

### 3. Access the Application
- **Web App**: http://localhost:8080
- **MySQL**: localhost:3306

### 4. Stop the Application
```bash
docker-compose down
```

## Manual Setup (Without Docker)

### Prerequisites
- Go 1.21+
- MySQL 8.0+

### 1. Install Dependencies
```bash
go mod download
```

### 2. Setup Database
```bash
# Create database
mysql -u root -p
CREATE DATABASE blogdb;
```

### 3. Run Application
```bash
go run main.go
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | 127.0.0.1 | Database host |
| `DB_PORT` | 3306 | Database port |
| `DB_USER` | root | Database user |
| `DB_PASSWORD` | dandan1234 | Database password |
| `DB_NAME` | blogdb | Database name |

## Project Structure

```
webapp/
├── handlers/          # HTTP handlers
│   ├── auth.go       # Authentication
│   └── post.go       # Post CRUD operations
├── middleware/        # Middleware functions
│   └── auth.go       # Session management
├── models/           # Data models
├── templates/        # HTML templates
├── static/          # Static files
├── database/        # Database configuration
├── utils/           # Utility functions
├── logs/            # Log files
├── Dockerfile       # Docker configuration
├── docker-compose.yml
└── main.go          # Application entry point
```

## Logging

The application creates detailed logs in the `logs/` directory:

- **`info.log`** - General application information
- **`error.log`** - Error messages and failures
- **`auth.log`** - Authentication events (login, logout, signup)

## Docker Services

### Web Application
- **Image**: Custom Go application
- **Port**: 8080
- **Features**: Ghost mode, authentication, CRUD operations

### MySQL Database
- **Image**: mysql:8.0
- **Port**: 3306
- **Database**: blogdb
- **Auto-initialization**: Tables and indexes created automatically

## Development

### View Logs
```bash
# View application logs
docker-compose logs webapp

# View database logs
docker-compose logs mysql

# Follow logs in real-time
docker-compose logs -f webapp
```

### Database Access
```bash
# Connect to MySQL container
docker-compose exec mysql mysql -u webapp -p blogdb
```

### Rebuild Application
```bash
# Rebuild and restart
docker-compose up --build webapp
```

## Security Features

- **Password Hashing** - bcrypt with cost factor 14
- **Secure Sessions** - HttpOnly cookies
- **Authorization** - Users can only edit/delete their own posts
- **Activity Logging** - Track all user actions and security events

## Ghost Mode Features

- **Floating Ghosts** - Animated ghosts on every page
- **Glowing Effects** - Matrix-style green glow animations
- **Spooky Messages** - Haunted error pages and logout messages
- **Loading Animations** - Ghost loading screens
- **Particle Effects** - Floating spooky emojis

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Home page with posts |
| GET | `/login` | Login page |
| POST | `/login` | Login authentication |
| GET | `/signup` | Signup page |
| POST | `/signup` | User registration |
| GET | `/logout` | Logout with spooky goodbye |
| GET | `/post/create` | Create post page |
| POST | `/post/create` | Create new post |
| GET | `/post/edit?id=X` | Edit post page |
| POST | `/post/edit` | Update post |
| GET | `/post/delete?id=X` | Delete post |

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check if MySQL is running
   docker-compose ps
   
   # Restart database
   docker-compose restart mysql
   ```

2. **Port Already in Use**
   ```bash
   # Change ports in docker-compose.yml
   ports:
     - "8081:8080"  # Use different port
   ```

3. **Permission Issues**
   ```bash
   # Fix log directory permissions
   sudo chown -R $USER:$USER logs/
   ```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test with Docker
5. Submit a pull request

## License

This project is open source and available under the MIT License.
