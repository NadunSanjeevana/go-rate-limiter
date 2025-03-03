# API Request Rate Limiter

## 📌 Project Overview
The **API Request Rate Limiter** is a middleware solution designed to control the rate of incoming API requests using **Leaky Bucket** and **Sliding Window** algorithms. It prevents abuse, ensures fair usage, and improves API reliability. The project also includes **JWT authentication**, **blacklist token management**, and **monitoring via Prometheus & Grafana**.

## 🚀 Features
- **Rate Limiting**: Implements request throttling using Leaky Bucket and Sliding Window algorithms.
- **JWT Authentication**: Secure API endpoints with JSON Web Tokens.
- **Blacklist Token Management**: Store revoked tokens in Redis to prevent reuse.
- **Redis Integration**: Efficient request counting and token storage.
- **Logging & Monitoring**: Tracks request rates and API performance using **Prometheus** & **Grafana**.

## 🛠️ Technologies Used
- **Go (Gin Framework)** - API development
- **Redis** - In-memory data store for rate limiting
- **JWT (JSON Web Tokens)** - Authentication
- **Prometheus & Grafana** - Monitoring and visualization
- **Docker & Docker Compose** - Containerized deployment

## 📂 Project Structure
```bash
📦 go-rate-limiter
 ┣ 📂 middleware        # Middleware for rate limiting and authentication
 ┣ 📂 pkg               # Package for Redis client
 ┣ 📂 utils             # JWT utilities (token generation & validation)
 ┣ 📂 config            # Configuration files (e.g., Prometheus, Docker Compose)
 ┣ 📄 main.go           # Main application entry point
 ┣ 📄 go.mod            # Go module dependencies
 ┣ 📄 README.md         # Project documentation
```

## 🔧 Installation & Setup
### 1️⃣ Clone the repository
```sh
git clone https://github.com/NadunSanjeevana/go-rate-limiter.git
cd go-rate-limiter
```

### 2️⃣ Install dependencies
Ensure you have **Go**, **Redis**, and **Docker** installed.
```sh
go mod tidy  # Install Go dependencies
```

### 3️⃣ Start Redis
```sh
redis-server
```

### 4️⃣ Run the API
```sh
go run main.go
```

### 5️⃣ Run with Docker
```sh
docker-compose up --build
```

## 📜 API Endpoints
### 🔐 Authentication
#### Login (Returns JWT Token)
```sh
POST /login
{
  "username": "user1",
  "role": "free"
}
```
#### Logout (Blacklist Token)
```sh
POST /logout
```

### 📌 Rate-Limited Routes
| Method | Endpoint  | Role-Based Limit |
|--------|----------|-----------------|
| GET    | `/ping`  | 5 (Free), 20 (Premium), 50 (Admin) |

## 📊 Monitoring with Prometheus & Grafana
### 1️⃣ Start Prometheus & Grafana
```sh
docker-compose up -d prometheus grafana
```
### 2️⃣ Access Dashboards
- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000` (Default login: `admin/admin`)

## 📌 Future Enhancements
- Add IP-based rate limiting.
- Implement a user-friendly dashboard for rate limit stats.
- Support for distributed systems with **Kafka**.

## 📜 License
This project is **open-source** under the MIT License.

## 📬 Contact
For any issues or contributions, reach out via [GitHub](https://github.com/NadunSanjeevana).
