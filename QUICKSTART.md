# GIN Authentication Microservice - Quick Start Guide

##  What You Have Now

After running the setup, your project contains:

```
GIN/
├── README.md                          # Comprehensive documentation
├── api-docs.html                      # Interactive API documentation
├── main.go                            # Updated with Swagger routes
├── go.mod                             # Go dependencies
├── .env                               # Environment variables
├── docs/
│   ├── swagger.yaml                   # OpenAPI 3.0 specification
│   ├── index.html                     # Swagger UI
│   └── schema.sql                     # Database schema (optional)
├── config/
│   ├── db.go                          # PostgreSQL connection
│   └── redis.go                       # Redis connection
├── controllers/
│   └── user.go                        # Auth handlers (23 endpoints)
├── middleware/
│   ├── auth.go                        # JWT validation
│   ├── admin-auth.go                  # Admin-only
│   ├── moderator-auth.go              # Moderator-only
│   ├── Activeate.go                   # User activation
│   ├── after-Change.go                # Post-change cleanup
│   ├── GeoGurd.go                     # Geolocation guard
│   ├── IsActive.go                    # Active status check
│   └── Reauth.go                      # Reauth verification
├── model/
│   └── user.go                        # Data models
├── routes/
│   └── user.go                        # Route definitions
└── utils/
    └── utils.go                       # Utilities (JWT, password, email)
```

---

##  Quick Start (5 Minutes)

### 1. Install Dependencies

```bash
cd /home/hosam/Desktop/GIN
go mod download
go mod tidy
```

### 2. Setup PostgreSQL Database

```bash
# Create database
psql -U postgres -c "CREATE DATABASE book;"

# Create tables
psql -U postgres -d book << 'EOF'
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

### 3. Setup Environment Variables

Create `.env` file:

```bash
cat > /home/hosam/Desktop/GIN/.env << 'EOF'
# PostgreSQL
HOST=localhost
PORT=5432
user_name=postgres
password=postgres
db_name=book
ssl=disable

# JWT
jwt_secret=this_is_my_secret_and_my_secret_alone_change_this_^_^
token_version=0

# Domain
domain=localhost

# Email (Gmail)
Mail_email=your-email@gmail.com
Mail_password=your-app-password

# Security
Pepper=this_is_your_king_change_this
Ms_API_KEY=your-mailersend-api-key

# OTP
opt_secret=this_is_my_secret_and_my_secret

# Password Reset
Gmail_Forget_password=http://localhost:3000/reset_password?token=
EOF
```

### 4. Start Redis

```bash
# macOS/Linux
redis-server

# Or in background
nohup redis-server &

# Docker
docker run -d -p 6379:6379 redis:latest
```

### 5. Run the Application

```bash
cd /home/hosam/Desktop/GIN
go run main.go
```

Expected output:
```
connected to database
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.
[GIN-debug] Loaded HTML Templates (1): qr_code.html
[GIN-debug] POST   /user/signup              --> GIN/controllers.Create (3 handlers)
[GIN-debug] POST   /user/login               --> GIN/controllers.Login (3 handlers)
...
[GIN-debug] Listening and serving HTTP on [::]:8080
```

### 6. Access Documentation

Open your browser:

- **Interactive API Docs:** http://localhost:8080/api-docs/
- **Alternative Docs:** http://localhost:8080/api-docs.html (standalone)
- **Swagger Spec:** http://localhost:8080/docs/swagger.yaml
- **Health Check:** http://localhost:8080/health

---

##  Documentation Files

### 1. **README.md** - Full Documentation
- Project overview
- Architecture diagram
- Security features explained
- Complete API documentation
- Database schema
- Authentication flows
- Installation guide
- Troubleshooting

### 2. **api-docs.html** - Standalone Viewer
- Tabbed interface (Overview, Features, Security, Database, Auth Flow, Endpoints, Setup, Testing)
- Feature cards
- Code examples
- All 23 endpoints documented
- Live testing section

### 3. **docs/index.html** - Interactive Swagger UI
- Swagger UI from CDN
- Try it out functionality
- Real-time API testing
- Request/response examples
- Schema documentation

### 4. **docs/swagger.yaml** - OpenAPI 3.0 Spec
- Complete API specification
- All 23 endpoints detailed
- Request/response schemas
- Authentication schemes
- Error codes
- Examples

---

##  23 API Endpoints

### Authentication (4)
- `POST /user/signup` - Register
- `POST /user/verify_signup_email` - Create account
- `POST /user/login` - Login
- `POST /user/verify-email` - Complete login

### 2FA (4)
- `POST /user/create-2fA` - Setup 2FA
- `POST /user/verify-2fA` - Verify 2FA code
- `POST /user/generte_login_codes` - Generate backup codes
- `POST /user/verify_login_code` - Login with backup code

### Password Management (4)
- `POST /user/forget_password` - Reset password
- `POST /user/reset_password` - Complete reset
- `POST /user/change-password` - Change password
- `POST /user/change-email` - Change email

### User Profile (2)
- `GET /user/me` - Get current user
- `GET /user/device-info` - Get device info

### Sessions (5)
- `POST /user/refresh` - Refresh token
- `GET /user/get-sessions` - List sessions
- `POST /user/logout` - Logout
- `POST /user/logout-all` - Logout all
- `POST /user/logout/:sessionid` - Logout specific

### Security (2)
- `POST /user/reauth-password` - Verify with password
- `POST /user/reauth-2fa` - Verify with 2FA

### Other (1)
- `POST /user/captcha-solved` - Mark CAPTCHA done

---

##   Testing Examples

### Test Signup
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

### Test Get User (with token)
```bash
curl -X GET http://localhost:8080/user/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Test Refresh Token
```bash
curl -X POST http://localhost:8080/user/refresh \
  -H "Cookie: refresh_token=YOUR_REFRESH_TOKEN"
```

---

##  Security Highlights

 **JWT Authentication:** 15-min access + 7-day refresh tokens
 **2FA Support:** TOTP with backup codes
 **Password Security:** Bcrypt+pepper, complexity checks, breach detection
 **Device Tracking:** Geolocation, browser info, last login
 **Rate Limiting:** CAPTCHA after 3 failed attempts, 15-min block after 5
 **Role-Based Access:** Admin, Moderator, User, Staff
 **Constant-Time Comparison:** Prevents timing attacks
 **Session Management:** Redis-backed with TTL
 **Reauth for Sensitive Ops:** Password/email changes require verification
 **Email Verification:** 6-digit codes with 15-min expiration

---

##  How to Use the Documentation

### For API Testing:
1. Go to http://localhost:8080/api-docs/
2. Scroll to any endpoint
3. Click "Try it out"
4. Fill in parameters
5. Click "Execute"
6. See live response

### For Reading Docs:
1. Read **README.md** for comprehensive overview
2. Open **api-docs.html** for tabbed documentation
3. Check specific endpoints in swagger.yaml

### For Integration:
1. Copy endpoint details from Swagger UI
2. Use Bearer token: `Authorization: Bearer {token}`
3. Follow request/response examples
4. Handle error codes as documented

---

## Common Issues & Fixes

### Database Connection Error
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Start PostgreSQL
brew services start postgresql  # macOS
sudo systemctl start postgresql  # Linux
```

### Redis Connection Error
```bash
# Check Redis is running
redis-cli ping

# Start Redis
redis-server
```

### Port Already in Use
```bash
# Find process using port 8080
lsof -i :8080

# Kill it
kill -9 <PID>
```

### Email Not Sending
```bash
# Use Gmail app-specific password (not regular password)
# Enable 2-Step Verification on Gmail account
# Check Mail_email and Mail_password in .env
```

---

## Project Structure Explanation

```
GIN/
├── main.go                     Entry point - starts server & routes
├── config/                     Database & Redis configuration
│   ├── db.go                   PostgreSQL connection setup
│   └── redis.go                Redis client initialization
├── routes/                     API route definitions
│   └── user.go                 All 23 endpoint routes
├── controllers/                Business logic handlers
│   └── user.go                 Authentication logic
├── middleware/                 Request processing
│   ├── auth.go                 JWT validation
│   ├── GeoGurd.go              Geographic restrictions
│   └── ...                     Other security middleware
├── model/                      Database models
│   └── user.go                 Data structures
├── utils/                      Helper functions
│   └── utils.go                JWT, password, email utilities
└── docs/                       API documentation
    ├── index.html              Swagger UI
    └── swagger.yaml            OpenAPI specification
```

---

##  Next Steps

1. **Customize Configuration:**
   - Update .env with your actual credentials
   - Change jwt_secret and Pepper to secure values
   - Configure email service

2. **Test All Endpoints:**
   - Use Swagger UI at http://localhost:8080/api-docs/
   - Try each endpoint with "Try it out"

3. **Read Full Documentation:**
   - Open README.md for detailed info
   - Check security features section
   - Understand authentication flows

4. **Deploy to Production:**
   - Follow production checklist in README.md
   - Use environment-specific .env files
   - Set up monitoring and logging
   - Configure HTTPS/TLS
   - Use strong secrets

---

##  Support

- **API Documentation:** http://localhost:8080/api-docs/
- **Code Documentation:** See README.md
- **Database Schema:** Check model/user.go
- **Environment Setup:** See .env.example

---

**Version:** 1.0.0
**Last Updated:** May 30, 2026
**Status:**  Production Ready
