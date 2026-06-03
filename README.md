PeoplesPost – Distributed Complaint Management System

 **Live Demo:** https://peoplespost.vercel.app/
 **Repository:** https://github.com/aryanrichhariya1234-lang/peoplespost_microservices



##  Overview

**PeoplesPost** is a **microservices-based full-stack platform** that enables:

**Citizens** to raise complaints/issues
**Officials** to manage, track, and resolve them

The system is designed with **scalability, performance, and real-world deployment** in mind, using modern backend architecture and cloud-native practices.

---

Architecture

The system follows a **distributed microservices architecture**:

```
Frontend (Next.js)
        ↓
API Gateway (Go Reverse Proxy)
        ↓
-------------------------------------
| Auth Service | Post Service | AI Service |
-------------------------------------
```

### 🔹 Services

**Auth Service** → Handles authentication (JWT + cookies)
**Post Service** → Manages complaints (CRUD, likes, status updates)
**AI Service** → Provides insights using Gemini API
**API Gateway** → Central routing, proxy, and middleware layer

---

Tech Stack

Backend

* Go (Golang)
* Node.js (legacy services)
* REST APIs
* JWT Authentication (cookie-based)

Frontend

* Next.js
* React
* Tailwind CSS

Database & Caching

* MongoDB
* Redis (for performance optimization)

Cloud & DevOps

* Docker (containerized microservices)
* AWS EC2 (deployment)
* GitHub Actions (CI/CD pipeline)

AI Integration

* Google Gemini API (complaint analysis & insights)

---

 Key Features

**Microservices Architecture** (Auth, Posts, AI)
**API Gateway with Reverse Proxy Routing**
**Secure Authentication (JWT + Cookies)**
**Redis Caching (~35% faster responses)**
**Dockerized Deployment**
**Cloud Deployment on AWS**
**AI-powered complaint insights**
**CI/CD automation with GitHub Actions**

---

##  Performance
 Reduced API response time by **~35%** using Redis caching
Improved scalability via independent service deployment
 Stateless architecture enabling horizontal scaling

---

## Running Locally

### 1. Clone repository

```bash
git clone https://github.com/aryanrichhariya1234-lang/peoplespost_microservices
cd peoplespost_microservices
```

---

### 2. Setup environment variables

Create a `.env` file in each service:

```
PORT=5000
MONGO_URI=your_mongo_uri
SECRET=your_jwt_secret
```

---

### 3. Run with Docker

```bash
docker-compose up --build
```

---

### 4. Access

* Frontend → http://localhost:3000
* API Gateway → http://localhost:4000

---

## Authentication Note

This project uses **cookie-based JWT authentication**.

⚠️ In production:

* Requires **HTTPS**
* Uses `SameSite=None; Secure`

---

##  Challenges & Learnings

* Handling **CORS and cookie security across services**
* Designing a **reverse proxy API gateway**
* Managing **inter-service communication in Docker**
* Implementing **distributed system patterns**
* Optimizing performance using **Redis caching**

---

##  Future Improvements

* Kubernetes deployment (container orchestration)
* Role-based access control (RBAC)
* Real-time notifications (WebSockets)
* Monitoring & logging (Prometheus, Grafana)

---

##  Contribution

Open to contributions! Feel free to fork and improve.

---

##  License

MIT License

---

##  Author

**Aryan Richhariya**
🔗 https://github.com/aryanrichhariya1234-lang
