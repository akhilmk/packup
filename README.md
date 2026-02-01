# PackUp

PackUp is a modern, collaborative task management application designed to bridge the gap between administrators and customers through a shared responsibility workflow.

## 🚀 Overview

PackUp allows teams to manage global requirements while giving individual users the flexibility to maintain their own personal checklists. The application features a clean, responsive interface and a robust backend designed for reliability and ease of deployment.

### Key Features

- **🛡️ Multi-Role System**: Distinct environments for Administrators and Customers.
- **🌍 Global Default Tasks**: Admins can push mandatory tasks to every user in the system.
- **🤝 Shared Responsibility**: Admins and Users can collaborate on shared tasks, tracking progress in real-time.
- **🔒 Privacy First**: Customers can keep personal tasks private or share them with admins for assistance.
- **⚡ Modern Tech Stack**: Built with Go, Svelte, PostgreSQL, and containerized with Docker.

---

## 🛠️ Tech Stack

- **Backend**: Go (Golang)
- **Frontend**: Svelte + Tailwind CSS
- **Database**: PostgreSQL
- **Infrastructure**: Docker & Docker Compose
- **Reverse Proxy**: Traefik (with automatic SSL)

---

## 🚦 Quick Start

### Prerequisites
- Docker & Docker Compose
- Make

### Local Development Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd packup
   ```

2. **Start the database**:
   ```bash
   make dev-build-up
   ```

3. **Build and Run the application**:
   ```bash
   make docker
   make run
   ```

The app will be available at [http://localhost:8080](http://localhost:8080).

---

## 📖 Documentation

For more detailed information, please refer to our documentation:

- **[Application Functionality](docs/requirements.md)**: Detailed breakdown of features, roles, and business logic.
- **[Developer Documentation](docs/developer-doc.md)**: A complete list of all available commands and development setup instructions.

---

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.
