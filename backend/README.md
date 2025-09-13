# Sudojo Backend - Multiplayer Sudoku Game

This is the backend application for Sudojo, a real-time multiplayer Sudoku game. The backend is built with Go, following a semi-hexagonal architecture pattern, and provides both HTTP and WebSocket APIs for real-time communication.

---

## Architecture Overview

The project follows a semi-hexagonal architecture with three main components:

- **Domain Layer**: Contains all business/game logic, rules, and models without any infrastructure dependencies. This layer is pure logic that can be tested independently.
- **Adapter Layer**: Connects external dependencies to the domain. (Note: This layer is not fully implemented in the current version)
- **Server Layer**: Orchestrates domain logic and hosts both HTTP and WebSocket endpoints for real-time communication.

---

## Architecture Rules

The project follows these dependency rules:

- **Domain** has no external dependencies
- **Adapter** depends only on domain
- **Server** can depend on both domain and adapters

This architecture enables:

- Clean separation of concerns
- Testable business logic
- Scalable and maintainable codebase
- Real-time multiplayer functionality

---

### Domain Layer

- **Purpose**: Encapsulates the business/game logic, rules, and models.
- **Examples**: `GameEngine`, `Player`, `Round`, `Scoring`, `GameRules`
- **Constraints**:
  - No direct knowledge of infrastructure
  - Pure logic, testable without external dependencies

### Adapter Layer

- **Purpose**: Connects external dependencies to the domain.
- **Examples**: `Database`, `Cache`, `MessageQueue`
- **Constraints**:
  - Implements interfaces defined in the domain
  - Infrastructure-specific code only

### Server Layer

- **Purpose**: Hosts both HTTP and WebSocket endpoints.
  - Orchestrates domain logic and adapters directly.
  - Manages player connections, lobby sessions, and real-time events.
- **Examples**:
  - `net/http` server setup
  - `gorilla/websocket` or native WebSocket handling
  - Routing and middleware registration
- **Constraints**:
  - Minimal business logic — delegates core game mechanics to domain
  - Can call domain and adapters directly

---

## Domain Logic

The domain layer implements the core game mechanics and contains the following components:

### Sudoku Game Engine

Located in `domain/sudoku.go`, this component handles:

- Sudoku board representation (9x9 grid)
- Move validation and execution
- Board state checking (completeness, validity)
- Puzzle generation with adjustable difficulty
- Solution verification

Key functionality:

- `ValidateMove`: Ensures moves follow Sudoku rules (row, column, box constraints)
- `MakeMove`: Applies valid moves to the board
- `IsComplete`: Checks if the puzzle is solved
- `GeneratePuzzle`: Creates new puzzles with a specified difficulty level
- `hasUniqueSolution`: Ensures generated puzzles have exactly one solution

### Lobby System

Located in `domain/lobby.go`, this component manages:

- Game lobby creation with unique IDs
- Puzzle and solution state tracking
- Random ID generation for lobby sharing

---

## API Endpoints and Communication

The backend provides the following endpoints and communication methods:

### HTTP Endpoints

- **Create Lobby**: `POST /lobby`
  - Creates a new game lobby
  - Returns: `{ id: string }` (Lobby ID)
- **Join Lobby**: `GET /lobby?id={lobbyId}` (WebSocket upgrade)
  - Joins an existing lobby and upgrades to WebSocket connection
  - Parameter: `id` - Lobby ID to join

### WebSocket Messages

#### Outgoing Messages (Client to Server)

- **Make Move**:

  ```json
  {
    "type": "move",
    "row": number,
    "col": number,
    "value": number
  }
  ```

- **Request Game State**:
  ```json
  {
    "type": "request_state"
  }
  ```

#### Incoming Messages (Server to Client)

- **Move Success**:

  ```json
  {
    "type": "success",
    "row": number,
    "col": number,
    "value": number,
    "success": true
  }
  ```

- **Move Error**:

  ```json
  {
    "type": "error",
    "success": false,
    "error": string
  }
  ```

- **Game State**:
  ```json
  {
    "type": "state",
    "board": number[][]
  }
  ```

---

## Server and WebSocket Implementation

The server layer (`server/server.go`) provides the following functionality:

### Connection Management

- Tracks active clients in each lobby
- Handles client connections and disconnections
- Broadcasts updates to appropriate clients
- Implements CORS for frontend-backend communication

### WebSocket Message Flow

1. Client connects to a lobby via WebSocket
2. Client can request current game state
3. Client can submit moves
4. Server validates moves using domain logic
5. Valid moves are broadcast to all connected clients
6. Invalid moves return error messages to the sender only

---

## Project Structure

```
backend/
├── domain/                # Domain layer with core game logic
│   ├── sudoku.go          # Sudoku game engine
│   ├── sudoku_test.go     # Tests for sudoku game engine
│   ├── lobby.go           # Lobby system implementation
│   └── lobby_test.go      # Tests for lobby system
├── server/                # Server layer
│   └── server.go          # HTTP and WebSocket server implementation
├── main.go                # Application entry point
└── go.mod                 # Go module definition
```

---

## Development Workflow

1. **Domain Logic Development**:

   - Implement core game logic in the domain layer
   - Ensure business rules are properly enforced
   - Write comprehensive tests for all domain components

2. **Server Implementation**:

   - Create HTTP endpoints for lobby management
   - Implement WebSocket handling for real-time communication
   - Connect domain logic to server endpoints

3. **Testing**:
   - Test domain components in isolation
   - Verify server functionality with integration tests
   - Ensure WebSocket communication works as expected

---

## Technical Decisions

- **Semi-Hexagonal Architecture**: Separates domain logic from infrastructure concerns
- **WebSockets for Real-time Communication**: Enables responsive multiplayer experience
- **Go for Backend**: Provides excellent concurrency for handling multiple game sessions
- **No External Dependencies in Domain**: Ensures testability and maintainability

---

## Implementation Guidelines

- **Code must be self-explanatory**
- **Use comments only were needed**
- **Variable naming rule**:
  - Short names for **local/temporary scope** (e.g., `i`, `p`, `g`)
  - Longer, descriptive names for **wider/larger scope** (e.g., `playerSession`, `gameEngine`)
- Favor **clear and expressive function names, struct names, and interface names** over comments
- Keep functions short and composable
- Maintain consistent **Go idioms** (package organization, error handling, naming)
- Prioritize **testability** by isolating logic in the domain
