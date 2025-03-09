# Schedulio.ai  - Schedule, post, and analyze your content seamlessly.

## Entry Point and Initialization
The main entry point of the application is in `main.go`, which initializes the database connection and starts the server.

### Key Functions:
- **Database Connection**: Establishes a connection to a PostgreSQL database using the `pgx` library.
- **Server Initialization**: Initializes a new server instance from the `api` package.

## Server Setup and Routing
The server is set up in `api/server.go`, where routes for user management and OAuth callbacks are defined.

### Key Routes:
- **User Registration**: `POST /account`
- **User Login**: `POST /account/login`
- **Get User**: `GET /account/:username`
- **Update User**: `PUT /account`

## Authentication Middleware
Authentication is handled in `api/authMiddleware.go`, which verifies tokens for protected routes.

### Key Functions:
- **Token Verification**: Checks for the presence of the authorization header and verifies the provided token.

## User Management
User data is managed through SQL queries defined in `db/queries/users.sql`, with a corresponding `User` model in `db/sqlc/models.go`.

### Key Queries:
- **Create User**: Inserts a new user into the database.
- **Get User**: Retrieves a user based on the username.
- **Update User**: Updates user information.
- **Delete User**: Deletes a user from the database.

## Token Handling
Token generation and verification are implemented in `tokens/paseto_auth.go`.

### Key Functions:
- **Create Token**: Generates a new token for a given username and duration.
- **Verify Token**: Decrypts a provided token and verifies its validity.