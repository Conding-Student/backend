# Golang Fiber v2 Project - FDSAP Intern Project Template

Welcome to the Golang Fiber v2 backend development project! This repository contains a backend service built using [Fiber v2](https://github.com/gofiber/fiber), a fast, lightweight, and flexible Go web framework tailored for modern backend development.

## Features

- **High Performance**: Built on top of Fasthttp for optimal speed.
- **RESTful API**: Clean and structured RESTful API design.
- **Middleware Support**: Preconfigured middleware for security, logging, and more.
- **Scalable Architecture**: Designed for scalability and maintainability.
- **Database Integration**: Supports PostgreSQL.

## Prerequisites

Before running the application, ensure you have the following installed:

- [Golang](https://golang.org/dl/) (version 1.18 or later)
- A database system (PostgreSQL)

## Installation

Follow these steps to set up and run the project locally:

1. Clone the repository:
   ```bash
   git clone https://github.com/FDSAP-Intern/golang_fiber_template.git
   ```
   then open your VS Code and open the project folder

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Configure environment variables:
   Create a `.env` file in the root directory and add the following:
   ```env
    DB_HOST = localhost
    DB_PORT = 5432
    DB_NAME = db_intern
    DB_UNME = fdsap.intern
    DB_PWRD = sh.intern@2025
    DB_TMEZ = asia/manila 
    DB_SSLM = disable
    PROJ_NAME = INTERN TEMPLATE V1
    PROJ_PORT = 5566
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

5. Access the application in your browser:
   ```
   http://localhost:5566
   ```

## Project Structure

```plaintext
.
├── main.go          # Entry point of the application
├── routes/          # API route definitions
├── controller/      # Business logic for each endpoint
├── model/           # Database models
├── middleware/      # Custom middleware
└── .env             # Environment variables
```

## Usage

### API Endpoints

| Method | Endpoint         | Description           |
|--------|------------------|-----------------------|
| GET    | `/api/v1/sample`  | Fetch all users       |
| POST   | `/api/v1/sample`  | Create a new user     |
| GET    | `/api/v1/sample/:id` | Fetch a user by ID  |

## Acknowledgments

- [Fiber Framework](https://github.com/gofiber/fiber)
- Golang Community

