# Mint Alternative

An open-source personal finance platform inspired by Mint, designed to make budgeting, expense tracking, and financial insights accessible to everyone.

Mint Alternative aims to provide a modern, extensible personal finance experience with a free core product, optional premium features, and support for multiple financial institution integrations.

> **Project Status:** Early development. The application is not yet ready for production use.

## Features

### Planned Core Features

- **Account Aggregation** — Connect bank accounts, credit cards, loans, and other financial accounts.
- **Transaction Tracking** — Automatically import and categorize transactions.
- **Budget Management** — Create monthly budgets and monitor spending against goals.
- **Financial Dashboard** — View account balances, spending trends, and overall financial health.
- **Net Worth Tracking** — Monitor assets and liabilities over time.
- **Recurring Transactions** — Identify subscriptions, bills, and recurring income.
- **Financial Goals** — Track savings targets and debt repayment progress.
- **Manual Accounts** — Track accounts without requiring a financial institution connection.

### Planned Integrations

The application is being designed with a provider-independent integration layer, allowing multiple financial data services to be supported.

Potential providers include:

- Plaid
- MX
- Finicity (Mastercard Open Finance)
- Other supported Open Banking providers

Provider availability, pricing, and supported features will vary. Integrations will be implemented incrementally.

## Technology Stack

The backend is being developed in Go, using a starter API as its initial foundation.

| Component | Technology |
|---|---|
| Backend | Go |
| API | REST / OpenAPI 3.1 |
| Database | PostgreSQL (Goose migrations, sqlc) |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS v4, shadcn/ui |
| Authentication | JWT (ECDSA ES256) & Argon2id |
| Financial Integrations | Provider-agnostic architecture |
| Deployment | Docker Compose / Cloud hosting (planned) |

The initial backend is based on a Go API starter project. Its existing architecture and dependencies will be evaluated and adapted as development progresses.

## Architecture

Mint Alternative is intended to use a modular architecture that separates financial data providers from core business logic.

```text
                  Frontend
                     |
                     v
                   Go API
                     |
       +-------------+-------------+
       |             |             |
    Accounts     Transactions    Budgets
       |             |             |
       +-------------+-------------+
                     |
              Data Access Layer
                     |
                PostgreSQL

         Financial Integration Layer
                     |
          +----------+----------+
          |          |          |
        Plaid        MX       Finicity
```

This approach is intended to support additional financial data providers without tightly coupling the application to a single service.

## Getting Started

The Go API uses PostgreSQL, JWT authentication, and River background jobs.

### Prerequisites

- Go (version determined by `go.mod`)
- Git
- PostgreSQL, when database persistence is configured
- Docker, if using containerized development

### Clone the Repository

```bash
git clone https://github.com/rpmoulton/mint-alternative.git
cd mint-alternative
```

### Install Dependencies

```bash
go mod download
```

### Run the Application

Copy `.env.example` to `.env` and review its local development settings. Then run:

```bash
make init
```

This installs development tools, generates local TLS and JWT keys, starts Docker
Compose, and applies application and River migrations. The API is available at
`https://localhost:3000` with a self-signed certificate. Use `make up` for later
starts. The example credentials are for local development only.

To run the API on the host with `make run`, export the variables from `.env` and
change the database URL hostnames from `database` to `localhost`.

Run `make test` to test the backend and `make openapi` to regenerate its API
contract.

## Development Roadmap

### Phase 1 — Foundation

- [x] Configure Go API starter
- [x] Establish database models and migrations
- [x] Implement user authentication
- [x] Implement accounts and transactions APIs
- [x] Configure local development environment

### Phase 2 — Personal Finance Essentials

- [x] Manual financial accounts
- [x] Transaction import and categorization
- [x] Budget creation and tracking
- [x] Financial dashboard
- [x] Net worth calculations

### Phase 3 — Financial Institution Integrations

- [ ] Define a common financial data provider interface
- [ ] Implement first banking integration
- [ ] Support account synchronization
- [ ] Handle duplicate transactions and reconciliation
- [ ] Add additional financial data providers

### Phase 4 — Advanced Features

- [ ] Recurring bill and subscription detection
- [ ] Savings and debt repayment goals
- [ ] Spending analytics and reporting
- [ ] Transaction rules and automation
- [ ] Data export and import

### Phase 5 — Sustainability

- [ ] Define optional premium features
- [ ] Evaluate nonintrusive advertising
- [ ] Establish infrastructure and integration cost controls
- [ ] Support long-term operation of the free tier

## Business Model

The goal is to keep essential personal finance functionality free.

Potential revenue sources include optional premium subscriptions, enhanced financial insights, and advertising where appropriate.

Core financial information should remain accessible without requiring a paid subscription.

## Security and Privacy

Because Mint Alternative will handle sensitive financial information, security and privacy are central design requirements.

Planned safeguards include:

- Encrypted communication using HTTPS
- Secure storage of financial integration credentials and tokens
- Least-privilege access to financial data
- Secure authentication and session management
- Protection against unauthorized account access
- User-controlled account disconnection and data deletion
- Avoiding the storage of bank login credentials wherever provider-hosted authorization is available

**Security Notice:** The project is under active development and has not been audited. Do not use it with real financial credentials or sensitive production data until appropriate security controls have been implemented and validated.

## Contributing

Contributions, suggestions, and feedback are welcome as the project develops.

Contribution guidelines and development standards will be added as the architecture stabilizes.

## License

License to be determined.

The project is intended to be open source. The final license will be selected after reviewing the licensing requirements of the starter repository and its dependencies.

## Acknowledgments

Mint Alternative is inspired by the original Mint personal finance application.

The backend is being built using an existing Go API starter as its foundation.

This project is independent and is not affiliated with Intuit or Mint.

---

**Mint Alternative — Personal finance without the unnecessary complexity.**