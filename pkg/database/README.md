# Database Support

This package provides unified database support for SQLite, PostgreSQL, and MySQL databases using GORM.

## Supported Databases

- **SQLite** - File-based database (default)
- **PostgreSQL** - Full-featured relational database
- **MySQL** - Popular relational database

## Configuration

### Environment Variables

You can configure the database using environment variables:

```bash
# Database type (sqlite, postgres, mysql)
export DB_TYPE=postgres

# Database connection details
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=your_password
export DB_NAME=lazy_ai_coder

# PostgreSQL specific
export DB_SSL_MODE=disable

# MySQL specific  
export DB_CHARSET=utf8mb4

# SQLite specific
export DB_FILE_PATH=lazy_ai_coder.db

# Default user credentials
export DEFAULT_USERNAME=admin
export DEFAULT_PASSWORD=admin123
export DEFAULT_EMAIL=admin@example.com
```

### Configuration File

Alternatively, you can configure the database in `config/config.yaml`:

```yaml
database:
  type: sqlite  # sqlite, postgres, mysql
  host: localhost
  port: 5432
  username: postgres
  password: ""
  database: lazy_ai_coder
  ssl_mode: disable
  charset: utf8mb4
  file_path: lazy_ai_coder.db
```

## Usage

### Initialize Database

```go
import "github.com/walterfan/lazy-ai-coder/pkg/database"

func main() {
    // Initialize database connection
    if err := database.InitDB(); err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer database.CloseDB()
    
    // Use the database
    db := database.GetDB()
    // ... your database operations
}
```

### Database Types

#### SQLite (Default)
```bash
export DB_TYPE=sqlite
export DB_FILE_PATH=lazy_ai_coder.db
```

#### PostgreSQL
```bash
export DB_TYPE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=your_password
export DB_NAME=lazy_ai_coder
export DB_SSL_MODE=disable
```

#### MySQL
```bash
export DB_TYPE=mysql
export DB_HOST=localhost
export DB_PORT=3306
export DB_USERNAME=root
export DB_PASSWORD=your_password
export DB_NAME=lazy_ai_coder
export DB_CHARSET=utf8mb4
```

## Docker Deployment

### PostgreSQL with Docker Compose

```yaml
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: your_password
      POSTGRES_DB: lazy_ai_coder
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  lazy-ai-coder:
    build: .
    environment:
      DB_TYPE: postgres
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USERNAME: postgres
      DB_PASSWORD: your_password
      DB_NAME: lazy_ai_coder
    depends_on:
      - postgres
```

### MySQL with Docker Compose

```yaml
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: your_password
      MYSQL_DATABASE: lazy_ai_coder
      MYSQL_USER: user
      MYSQL_PASSWORD: user_password
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  lazy-ai-coder:
    build: .
    environment:
      DB_TYPE: mysql
      DB_HOST: mysql
      DB_PORT: 3306
      DB_USERNAME: user
      DB_PASSWORD: user_password
      DB_NAME: lazy_ai_coder
    depends_on:
      - mysql
```

## Features

- **Auto-migration**: Automatically creates and updates database schema
- **Environment-based configuration**: Override settings with environment variables
- **Default data initialization**: Creates default admin user and loads prompts
- **Connection pooling**: Managed by GORM
- **Error handling**: Comprehensive error handling and logging

## Migration from SQLite

To migrate from SQLite to PostgreSQL or MySQL:

1. Export your SQLite data
2. Update your configuration
3. Import data to the new database
4. Restart the application

The application will automatically handle schema migration and data initialization. 