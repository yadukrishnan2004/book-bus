book-bus/
├── cmd/
│   └── api/
│       └── main.go          # Entry point for the Gin server
├── internal/                # Private application code
│   ├── handlers/            # HTTP request handlers (controllers)
│   ├── models/              # Data models (Bus, Seat, Booking, User)
│   ├── repository/          # Database queries & transactions
│   └── services/            # Core business logic (e.g., booking validation)
├── db/                      # Database migrations / SQL scripts
├── Dockerfile.dev           # Dockerfile tailored for local development (hot-reload)
├── docker-compose.yml       # Orchestrates Go app + PostgreSQL database
├── .env                     # Environment variables (DB credentials, secret keys)
├── .gitignore
├── go.mod
└── go.sum


