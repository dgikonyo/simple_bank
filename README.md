# **Project: Simple Bank - Dev Guide**
---
## **1. Project Structure**

We follow the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) to keep code organized and scalable.

```text
simple-bank/
├── cmd/
│   └── api/
│       └── main.go           # Entry point for the application
├── internal/
│   └── db/                   # Auto-generated SQLC code (Go models & queries)
├── sql/
│   ├── queries/              # .sql files containing your CRUD queries
│   └── schema/               # .sql files containing your CREATE TABLE statements
├── Dockerfile                # Multi-stage build (Dev & Production)
├── docker-compose.yml        # Orchestrates Go API & Postgres DB
├── Makefile                  # Project automation shortcuts
├── sqlc.yaml                 # SQLC configuration
├── .air.toml                 # Hot-reload configuration
└── go.mod                    # Go module definition

```

---

## **2. Getting Started**

Follow these steps to get your environment running from scratch:

### **Step 1: Initialize the Project**

Run the initialization command to set up your Go modules.

```bash
make init

```

### **Step 2: Database Code Generation**

Define your tables in `sql/schema/` and your queries in `sql/queries/`, then run:

```bash
make gen

```

*This will populate `internal/db/` with type-safe Go code.*

### **Step 3: Start Development Environment**

Spin up the database and the backend with hot-reloading enabled:

```bash
make dev

```

*The API will be available at `http://localhost:8080`.*

---

## **3. Common Commands (Makefile)**

| Command | Description |
| --- | --- |
| `make init` | Initializes Go modules and tidies dependencies. |
| `make gen` | Runs SQLC to generate Go code from SQL files. |
| `make dev` | Builds and starts all containers with live-reloading. |
| `make stop` | Stops and removes active containers. |
| `make clean` | Wipes database volumes and deletes generated code for a fresh start. |

---

## **4. Database Configuration**

The PostgreSQL instance is managed by Docker. You can connect to it using your favorite DB client (DBeaver, TablePlus, etc.):

* **Host:** `localhost`
* **Port:** `5432`
* **User:** `root_user`
* **Password:** `root_secret`
* **Database:** `bank_db`

---

## **5. Developer Workflow**

1. **Modify Schema:** Edit/Add files in `sql/schema/`.
2. **Modify Queries:** Edit/Add files in `sql/query/`.
3. **Sync Code:** Run `make gen`.
4. **Write Logic:** Use the new methods in `internal/db` within your handlers in `cmd/api/`.
5. **Save:** The `Air` watcher in Docker will automatically re-compile and restart your app.

---

## **6. Production**

To build the optimized, lightweight production image (without Air and the Go SDK):

```bash
docker build --target runner -t simple-bank:latest .

```

---