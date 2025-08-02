# GIN Authentication Microservice

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [Security Features](#security-features)
- [Database Schema](#database-schema)
- [API Documentation](#api-documentation)
- [Installation & Setup](#installation--setup)
- [Running the Application](#running-the-application)
- [API Testing](#api-testing)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Production Deployment](#production-deployment)

---

## Project Overview

This is an **enterprise-grade authentication microservice** built with:
- **Go** with **Gin Web Framework** for high-performance HTTP handling
- **PostgreSQL** for persistent data storage
- **Redis** for session management and caching
- **JWT** for stateless authentication
- **2FA (TOTP)** for enhanced security
- **Role-Based Access Control** for authorization

### Key Features

- Email & password authentication with verification
- Two-Factor Authentication (2FA) with TOTP & backup codes
- JWT-based token management with refresh tokens
- Password reset with secure token flow
- Geolocation-based security (country blocking)
- Device tracking and session management
- Password strength analysis with zxcvbn
- Password breach checking (HIBP integration)
- Account reauthorization for sensitive operations
- Password history tracking (prevent reuse)
- Rate limiting & brute force protection
- Role-based access (Admin, Moderator, User, Staff)
- Constant-time comparison for security
- Comprehensive error handling

---

## Architecture

### Layered Architecture

```
┌─────────────────────────────────────────────────┐
│           HTTP Client Requests                   │
└────────────────────┬────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────┐
│         Gin HTTP Server (Port 8080)              │
│  Routes → Middleware → Controllers → Models     │
└─────────────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
   ┌─────────┐ ┌──────────┐ ┌──────────┐
   │ Auth    │ │ Geo      │ │ Session  │
   │ Middleware│ │ Guard    │ │Management│
   └─────────┘ └──────────┘ └──────────┘
        │            │            │
        └────────────┬────────────┘
                     ▼
        ┌──────────────────────┐
        │   Business Logic     │
        │   (Controllers)      │
        └──────────────────────┘
                     │
        ┌───────────┴───────────┐
        ▼                       ▼
   ┌─────────────┐       ┌──────────┐
   │  PostgreSQL │       │  Redis   │
   │  Database   │       │  Cache   │
   └─────────────┘       └──────────┘
```

### Module Structure

- `routes/` - API endpoint definitions and routing
- `controllers/` - Business logic for authentication endpoints
- `middleware/` - Authentication, authorization, security guards
- `model/` - Database models (Users, DeviceRecord, OldPassword, Session)
- `config/` - Database and Redis connection configuration
- `utils/` - JWT generation, password hashing, email sending, validation
- `docs/` - API documentation (Swagger/OpenAPI)

---

## Security Features

### 1. Password Security

**Storage**:
- Algorithm: `bcrypt` with cost factor 12
- Pepper: Additional secret appended before hashing
- Format: `$2a$12$...` (120+ characters)

**Validation**:
- Length: 12-128 characters
- Complexity:
  - 3+ uppercase letters
  - 3+ lowercase letters
  - 3+ numbers
  - 3+ special symbols
- Breach checking against Have I Been Pwned
- Strength analysis with zxcvbn algorithm

**History**:
- Last 5 passwords stored
- Prevents recent password reuse
- Checked on password change

### 2. JWT Security

**Token Structure**:
```
Header:
{
  "alg": "HS256",
  "typ": "JWT"
}

Payload:
{
  "Username": "john_doe_a1b2c3d4",
  "email": "john@gmail.com",
  "role": "user",
  "id": 1,
  "exp": 1699564800,
  "iat": 1699564500,
  "version": 0,
  "jti": "uuid-string",
  "devid": 1
}
```

**Validation**:
- Signature verification with HMAC-SHA256
- Expiration check (exp claim)
- Blocklist check in Redis for revoked tokens
- Token version matching (invalidates on password change)

**Token Types**:
- Access Token: 15 minutes validity
- Refresh Token: 7 days validity (HTTP-only cookie)

### 3. Two-Factor Authentication (2FA)

**TOTP Implementation**:
- RFC 6238 standard
- 30-second time window
- QR code generation for authenticator apps
- 12 backup login codes provided

**2FA Flow**:
1. User logs in with email/password
2. System checks if 2FA enabled
3. Sends verification code or asks for authenticator code
4. After verification, issues access token

### 4. Rate Limiting & Brute Force Protection

**Login Attempt Tracking**:
```
1-2 attempts  → Allow retry
3 attempts    → CAPTCHA required
4 attempts    → CAPTCHA + warning
5+ attempts   → 15-minute block
```

**Redis Keys**:
- `Login:fail:{email}` - Current attempt count
- `Login:block:{email}` - Block flag (TTL: 15 min)
- `captcha:passed:{email}` - CAPTCHA completion flag

### 5. Geolocation-Based Security

**Features**:
- IP-based geolocation using ip-api.com
- Blocks restricted countries (Ukraine, Russia - configurable)
- Location consistency checks
- Detects impossible travel/account compromise

**Applied On**:
- User signup
- User login
- Every authenticated request

### 6. Device Tracking & Session Management

**Device Record Storage**:
- Unique device ID
- Geographic location (city, region, country, coordinates)
- Browser/User-Agent
- Last login timestamp
- GPS coordinates (latitude/longitude)

**Session Management**:
```
Redis Key: "session:{userId}:{jti}"

Data:
{
  "Jti": "uuid-string",
  "UserID": 1,
  "IsActive": true,
  "IssuedAT": "2024-01-01T00:00:00Z",
  "DeviceInfoId": 1,
  "ExpireAt": "2024-01-01T00:15:00Z"
}
```

### 7. Constant-Time Comparison

Prevents timing attacks on:
- Verification codes
- Password comparisons
- Token validation

### 8. Email Verification

- 6-digit codes sent via SMTP
- 15-minute expiration
- Constant-time comparison
- Failed signup doesn't create user until verified

### 9. Security Headers

- No sensitive data in logs
- Error messages don't reveal user existence
- SQL injection prevention via GORM ORM
- XSS prevention through response encoding

---

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Authentication
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(320) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    token_version INT DEFAULT 0,
    
    -- Email Verification
    is_email_verified BOOLEAN DEFAULT FALSE,
    verification_code VARCHAR(255),
    verification_expires_at TIMESTAMP WITH TIME ZONE,
    
    -- 2FA & Security
    tfa_code VARCHAR(255),
    tfa_verified BOOLEAN DEFAULT FALSE,
    login_codes TEXT,
    login_codes_set BOOLEAN DEFAULT FALSE,
    
    -- User Profile
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) UNIQUE,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female')),
    birthday DATE,
    
    -- Roles & Status
    role VARCHAR(20) CHECK (role IN ('admin', 'user', 'moderator', 'staff')) DEFAULT 'user',
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    
    -- Profile Content
    profile_image_url VARCHAR(500),
    cover_image_url TEXT,
    bio TEXT,
    
    -- OAuth Support
    provider VARCHAR(50)
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
```

### Device Record Table

```sql
CREATE TABLE device_record (
    id SERIAL PRIMARY KEY,
    userid INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Geographic Information
    city VARCHAR(255),
    region VARCHAR(255),
    country VARCHAR(255),
    locale VARCHAR(10),
    lat FLOAT8,
    lon FLOAT8,
    zipcode VARCHAR(50),
    
    -- Device Information
    browser TEXT,
    last_login TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_device_record_userid ON device_record(userid);
```

### Old Password Table

```sql
CREATE TABLE old_password (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_old_password_user_id ON old_password(user_id);
```

### Data Model Relationships

```
Users (1) ──── (M) DeviceRecord
   │
   └──── (M) OldPassword
```

---

## API Documentation

### Base URL
```
http://localhost:8080/user
```

### Authentication
```
Authorization: Bearer {access_token}
Cookie: refresh_token={refresh_token}
```

### Response Format

**Success (200/202)**:
```json
{
  "field": "value"
}
```

**Error (4xx/5xx)**:
```json
{
  "error": "Error message"
}
```

---

### **1. Signup** - Register new user

**Endpoint**: `POST /user/signup`

**Access**: Public (GeoGuard applied)

**Request**:
```json
{
  "name": "John Doe",
  "email": "john@gmail.com",
  "password": "SecurePass123!@#",
  "phone": "1234567890",
  "gender": "male",
  "ZipCode": "12345"
}
```

**Validation**:
- `name`: 2-255 chars, letters/spaces only
- `email`: Valid Gmail format, not registered
- `password`: 12-128 chars with complexity rules
- `phone`: Numeric only
- `gender`: "male" or "female"
- `ZipCode`: Required

**Response (202)**:
```json
{
  "message": "Verify your email to continue"
}
```

**Errors**:
- 400: Invalid email or weak password
- 401: Geolocation blocked
- 500: Server error

---

### **2. Verify Signup Email** - Create account

**Endpoint**: `POST /user/verify_signup_email`

**Request**:
```json
{
  "email": "john@gmail.com",
  "code": "123456"
}
```

**Response (200)**:
```json
{
  "user": {
    "id": 1,
    "username": "johndoe_a1b2c3d4",
    "email": "john@gmail.com",
    "name": "John Doe",
    "role": "user"
  }
}
```

---

### **3. Login** - Authenticate user

**Endpoint**: `POST /user/login`

**Request**:
```json
{
  "email": "john@gmail.com",
  "password": "SecurePass123!@#"
}
```

**Response (200)**:
```json
{
  "message": "Enter verification code to continue"
}
```

---

### **4. Verify Email (Login)** - Complete login

**Endpoint**: `POST /user/verify-email`

**Request**:
```json
{
  "email": "john@gmail.com",
  "code": "123456"
}
```

**Response (200)** - Without 2FA:
```json
{
  "user": {...},
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Response (200)** - With 2FA Enabled:
```json
{
  "message": "Enter your 2FA code to login"
}
```

---

### **5. Create 2FA** - Setup two-factor authentication

**Endpoint**: `POST /user/create-2fA`

**Request**:
```json
{
  "email": "john@gmail.com"
}
```

**Response (200)**:
```json
{
  "email": "john@gmail.com",
  "png": "data:image/png;base64,iVBORw0KGgo...",
  "secret": "JBSWY3DPEBLW64TMMQ6AAAAA"
}
```

Returns QR code and secret for authenticator app scanning.

---

### **6. Verify 2FA** - Authenticate with 2FA code

**Endpoint**: `POST /user/verify-2fA`

**Request**:
```json
{
  "email": "john@gmail.com",
  "code": "123456"
}
```

**Response (200)**:
```json
{
  "user": {...},
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

### **7. Generate Backup Codes** - Create recovery codes

**Endpoint**: `POST /user/generte_login_codes`

**Request**:
```json
{
  "email": "john@gmail.com"
}
```

**Response (200)**:
```json
{
  "message": "Backup codes generated and sent to email"
}
```

Generates 12 one-time recovery codes.

---

### **8. Verify Backup Code** - Login with backup code

**Endpoint**: `POST /user/verify_login_code`

**Request**:
```json
{
  "email": "john@gmail.com",
  "code": "123456"
}
```

**Response (200)**:
```json
{
  "user": {...},
  "token": "...",
  "refresh_token": "..."
}
```

---

### **9. Refresh Token** - Get new access token

**Endpoint**: `POST /user/refresh`

**Response (200)**:
```json
{
  "NewAcesstoken": "eyJhbGciOiJIUzI1NiIs..."
}
```

Requires valid refresh token cookie.

---

### **10. Forget Password** - Request password reset

**Endpoint**: `POST /user/forget_password`

**Request**:
```json
{
  "email": "john@gmail.com"
}
```

**Response (200)**:
```json
{
  "message": "Password reset link sent to email"
}
```

---

### **11. Reset Password** - Complete password reset

**Endpoint**: `POST /user/reset_password`

**Request**:
```json
{
  "email": "john@gmail.com",
  "token": "uuid-string",
  "password": "NewSecurePass123!@#"
}
```

**Response (200)**:
```json
{
  "message": "Password reset successfully"
}
```

---

### **12. Change Password** - Authenticated password change

**Endpoint**: `POST /user/change-password`

**Access**: Authenticated + Reauth verified

**Request**:
```json
{
  "password": "NewSecurePass123!@#",
  "confirm": "NewSecurePass123!@#"
}
```

**Response (200)**:
```json
{
  "message": "Password updated",
  "score": {
    "Score": 3,
    "Entropy": 85.5,
    "CrackTime": 1000000000
  }
}
```

Forces re-login after change.

---

### **13. Change Email** - Initiate email change

**Endpoint**: `POST /user/change-email`

**Access**: Authenticated + Reauth verified

**Request**:
```json
{
  "email": "newemail@gmail.com"
}
```

**Response (200)**:
```json
{
  "message": "Verify new email to complete change"
}
```

---

### **14. Verify New Email** - Confirm email change

**Endpoint**: `POST /user/verify-new-email`

**Request**:
```json
{
  "email": "newemail@gmail.com",
  "code": "123456"
}
```

**Response (200)**:
```json
{
  "message": "Email changed successfully"
}
```

---

### **15. Reauth with Password** - Verify for sensitive ops

**Endpoint**: `POST /user/reauth-password`

**Access**: Authenticated

**Request**:
```json
{
  "email": "john@gmail.com",
  "password": "SecurePass123!@#"
}
```

**Response (200)**:
```json
{
  "message": "You can now change your credentials"
}
```

Sets 5-minute reauth window.

---

### **16. Reauth with 2FA** - Verify with 2FA

**Endpoint**: `POST /user/reauth-2fa`

**Access**: Authenticated

**Request**:
```json
{
  "email": "john@gmail.com",
  "code": "123456"
}
```

**Response (200)**:
```json
{
  "message": "You can now change your credentials"
}
```

---

### **17. Get Current User** - Retrieve profile

**Endpoint**: `GET /user/me`

**Access**: Authenticated

**Response (200)**:
```json
{
  "user": {
    "id": 1,
    "username": "johndoe_a1b2c3d4",
    "email": "john@gmail.com",
    "name": "John Doe",
    "phone": "1234567890",
    "gender": "male",
    "role": "user",
    "is_email_verified": true,
    "is_active": true,
    "tfa_verified": false,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### **18. Get Device Info** - Retrieve device details

**Endpoint**: `GET /user/device-info?email=john@gmail.com`

**Response (200)**:
```json
{
  "id": 1,
  "user_id": 1,
  "city": "San Francisco",
  "region": "California",
  "country": "United States",
  "locale": "en-US",
  "lat": 37.7749,
  "lon": -122.4194,
  "browser": "Mozilla/5.0...",
  "last_login": "2024-01-01T12:30:00Z"
}
```

---

### **19. Get Sessions** - List active sessions

**Endpoint**: `GET /user/get-sessions`

**Access**: Authenticated

**Response (200)**:
```json
[
  {
    "Jti": "uuid-1",
    "UserID": 1,
    "IsActive": true,
    "IssuedAT": "2024-01-01T00:00:00Z",
    "DeviceInfoId": 1,
    "ExpireAt": "2024-01-01T00:15:00Z"
  }
]
```

---

### **20. Logout** - Close current session

**Endpoint**: `POST /user/logout`

**Access**: Authenticated

**Response (200)**:
```json
{
  "message": "Logged out successfully"
}
```

---

### **21. Logout All Sessions** - Close all sessions

**Endpoint**: `POST /user/logout-all`

**Access**: Authenticated

**Response (200)**:
```json
{
  "message": "Logged out from all sessions"
}
```

---

### **22. Logout Specific Session** - Close specific session

**Endpoint**: `POST /user/logout/:sessionid`

**Access**: Authenticated

**Response (200)**:
```json
{
  "message": "Session closed"
}
```

---

### **23. CAPTCHA Solved** - Mark CAPTCHA completion

**Endpoint**: `POST /user/captcha-solved?email=john@gmail.com`

**Response (200)**:
```json
{
  "message": "CAPTCHA verified"
}
```

Bypasses rate limiting for next login attempt.

---

## Installation & Setup

### Prerequisites

- **Go**: 1.23+ ([Download](https://golang.org/dl/))
- **PostgreSQL**: 12+ ([Download](https://www.postgresql.org/download/))
- **Redis**: 6.0+ ([Download](https://redis.io/download))
- **Git**: For version control

### Step 1: Clone Repository

```bash
cd /home/hosam/Desktop
git clone https://github.com/hossamelden202/GIN_AUTH.git
cd GIN
```

### Step 2: Install Dependencies

```bash
go mod download
go mod tidy
```

### Step 3: PostgreSQL Setup

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE book;

# Exit
\q
```

### Step 4: Create Tables

```bash
psql -U postgres -d book << EOF
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    username VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(320) UNIQUE NOT NULL,
    is_email_verified BOOLEAN DEFAULT FALSE,
    verification_code VARCHAR(255),
    verification_expires_at TIMESTAMP WITH TIME ZONE,
    phone VARCHAR(20) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    profile_image_url VARCHAR(500),
    cover_image_url TEXT,
    bio TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    gender VARCHAR(10),
    birthday DATE,
    is_active BOOLEAN DEFAULT TRUE,
    tfa_verifed BOOLEAN DEFAULT FALSE,
    login_codes TEXT,
    login_codes_set BOOLEAN DEFAULT FALSE,
    tfa_code VARCHAR(255),
    token_version INT DEFAULT 0,
    provider VARCHAR(50)
);

CREATE TABLE device_record (
    id SERIAL PRIMARY KEY,
    userid INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    city VARCHAR(255),
    region VARCHAR(255),
    country VARCHAR(255),
    locale VARCHAR(10),
    lat FLOAT8,
    lon FLOAT8,
    zipcode VARCHAR(50),
    browser TEXT,
    last_login TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE old_password (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_device_record_userid ON device_record(userid);
CREATE INDEX idx_old_password_user_id ON old_password(user_id);
EOF
```

### Step 5: Environment Variables

Create `.env` file in project root:

```bash
cat > .env << EOF
# Database Configuration
HOST=localhost
PORT=5432
user_name=postgres
password=your_password
db_name=book
ssl=disable

# JWT Configuration
jwt_secret=this_is_my_secret_and_my_secret_alone_^_^
token_version=0

# Domain Configuration
domain=localhost

# Email Configuration (Gmail SMTP)
Mail_email=your-email@gmail.com
Mail_password=your-app-specific-password

# Security
Pepper=this_is_your_king
Ms_API_KEY=your-mailersend-api-key

# OTP Configuration
opt_secret=this_is_my_secret_and_my_secret

# Password Reset URL
Gmail_Forget_password=http://localhost:3000/reset_password?token=
EOF
```

### Step 6: Start Redis

```bash
# macOS/Linux
redis-server

# Windows
redis-server.exe

# Docker
docker run -d -p 6379:6379 redis:latest
```

### Step 7: Install Go Packages

```bash
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
go get github.com/google/uuid
go get github.com/joho/godotenv
go get gorm.io/driver/postgres
go get gorm.io/gorm
go get github.com/redis/go-redis/v9
go get golang.org/x/crypto
go get github.com/skip2/go-qrcode
go get github.com/pquerna/otp
go get github.com/nbutton23/zxcvbn-go
```

---

## Running the Application

### Development Mode

```bash
go run main.go
```

Expected output:
```
connected to database
[GIN-debug] Listening and serving HTTP on [::]:8080
```

### Production Mode

```bash
# Build
go build -o gin-auth-server main.go

# Run
./gin-auth-server
```

### Docker Deployment

```bash
# Build image
docker build -t gin-auth:latest .

# Run container
docker run -p 8080:8080 --env-file .env gin-auth:latest
```

---

## API Testing

### Using cURL

**Signup**:
```bash
curl -X POST http://localhost:8080/user/signup \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@gmail.com",
    "password": "SecurePass123!@#Test",
    "phone": "1234567890",
    "gender": "male",
    "ZipCode": "94105"
  }'
```

**Login**:
```bash
curl -X POST http://localhost:8080/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@gmail.com",
    "password": "SecurePass123!@#Test"
  }'
```

**Get Current User**:
```bash
curl -X GET http://localhost:8080/user/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Using Postman

1. Open Postman
2. Import → File → Select `docs/swagger.json`
3. Automatically creates collection with all endpoints
4. Set Authorization header with Bearer token

### Using Swagger UI

Open browser to: `http://localhost:8080/api-docs/`

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `HOST` | PostgreSQL host | localhost |
| `PORT` | PostgreSQL port | 5432 |
| `user_name` | DB username | postgres |
| `password` | DB password | required |
| `db_name` | Database name | book |
| `ssl` | SSL mode | disable |
| `jwt_secret` | JWT signing secret | required |
| `domain` | Cookie domain | localhost |
| `Mail_email` | Sender email | required |
| `Mail_password` | Email password | required |
| `Pepper` | Password hash pepper | required |
| `opt_secret` | OTP secret | required |

### Security Configuration

**For Production**:
```bash
# Generate strong JWT secret (32+ chars)
openssl rand -base64 32

# Generate strong pepper (24+ chars)
openssl rand -base64 24

# Update .env
domain=yourdomain.com
ssl=require
```

---

## Troubleshooting

### Database Connection Error

```
error connecting database: connection refused
```

**Solution**: Ensure PostgreSQL is running
```bash
# macOS
brew services start postgresql

# Linux
sudo systemctl start postgresql

# Windows
# Start PostgreSQL service
```

### Redis Connection Error

```
something went wrong during connection database
```

**Solution**: Start Redis server
```bash
redis-server
```

### Email Sending Fails

```
something went wrong: {smtp error}
```

**Solution**:
- Use Gmail app-specific password
- Enable 2-Step Verification on Gmail
- Check `Mail_email` and `Mail_password` in `.env`

### Password Validation Error

```
password should contain atleast 3 lower case char
```

**Solution**: Use password with required complexity:
- 12+ characters
- 3+ uppercase
- 3+ lowercase
- 3+ numbers
- 3+ special characters

### Geolocation Block

```
users in this country cannot access my website
```

**Solution**: Edit `middleware/GeoGurd.go` to modify blocked countries

---

## Production Deployment Checklist

- [ ] Change `domain` to actual domain
- [ ] Set `ssl=require` in PostgreSQL
- [ ] Generate strong `jwt_secret` (32+ chars)
- [ ] Generate strong `Pepper` (24+ chars)
- [ ] Configure HTTPS/TLS certificates
- [ ] Set secure cookie flags
- [ ] Configure CORS for allowed origins
- [ ] Set up email service (Gmail, SendGrid, etc.)
- [ ] Configure database backups
- [ ] Set up Redis persistence
- [ ] Configure rate limiting appropriately
- [ ] Test all security features
- [ ] Set up CI/CD pipeline
- [ ] Configure environment-specific settings
- [ ] Set up error tracking (Sentry, etc.)
- [ ] Configure log aggregation
- [ ] Review security audit

---

## Support

- Issues: Open GitHub issue
- Documentation: See `/docs` folder
- API Docs: http://localhost:8080/api-docs/
- Email: Contact development team

---

## License

MIT License - See LICENSE file for details

---

## Authors

- Security Team - Authentication & Authorization
- Backend Team - API Development
- DevOps Team - Deployment & Infrastructure

---

**Last Updated**: May 30, 2026
**Version**: 1.0.0
