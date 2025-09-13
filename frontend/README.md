# Sudojo Frontend - Multiplayer Sudoku Game

This is the frontend application for Sudojo, a real-time multiplayer Sudoku game. The frontend is built with React (v19.1.1) and TypeScript, using Vite (v7.1.2) as the build tool. It communicates with the backend via HTTP and WebSocket APIs.

---

## Architecture Overview

The frontend follows a component-based architecture with an emphasis on:

- **Isolated Components**: Components are designed to be generic and reusable, avoiding tight coupling to specific business logic.
- **Custom Hooks**: All business logic and communication with the backend is encapsulated in custom hooks.
- **Prop Drilling**: State is passed down through props rather than using global state management.
- **Clean Separation**: UI components are separated from business logic.

---

## User Experience

The frontend aims to provide a zen-like experience for users:

- **Minimal Distraction**: Clean interface that keeps users focused on the game without unnecessary elements.
- **Glassmorphism Design**: Utilizing glassmorphism as the primary visual theme for a modern, elegant look.
- **Smooth Interactions**: Carefully designed animations and transitions that feel natural and unobtrusive.
- **Responsive Feedback**: Subtle visual and audio cues that provide feedback without breaking immersion.
- **Adaptive Layout**: Seamless experience across different device sizes while maintaining the zen aesthetic.
- **Visual Consistency**: Maintaining cohesive visual language, spacing, and interaction patterns throughout the application to create a harmonious and predictable experience.

The goal is to create an environment where players can enjoy the intellectual challenge of Sudoku while feeling calm and focused.

---

## Component Guidelines

### Generic Components

Components should be designed to be as generic as possible:

- **Prefer Composition**: Build specific components by composing generic ones.
- **Parameterize Behavior**: Use props to modify component behavior rather than creating specialized components.
- **Example**: Use a generic `Button` component with customizable props instead of creating a specialized `PlayButton`.

```tsx
// ❌ Avoid specialized components
const PlayButton = () => <button className="play-button">Play</button>;

// ✅ Use generic components with props
const Button = ({ children, variant, ...props }) => (
  <button className={`button button-${variant}`} {...props}>
    {children}
  </button>
);

// Usage
<Button variant="play">Play</Button>;
```

### State Management

- **Avoid Global State**: Global state should be avoided in favor of local component state and prop drilling.
- **Props Drilling**: Pass state down through component props, even if it requires passing through multiple levels.
- **Context**: Only use React Context for truly global concerns like theming or authentication, not for application state.

### Component Testing

All components should be designed to be testable in isolation:

- **Testing Framework**: Vitest (v3.2.4) is used as the testing framework.
- **Dependency Injection**: Pass dependencies as props rather than importing them directly.
- **Testable Props**: Ensure all behavior can be controlled via props.
- **Mock APIs**: Design components to work with mock API implementations for testing.
- **Colocated Tests**: Test files should be placed in the same directory as the file they are testing.

---

## Custom Hooks

All business logic and API communication must be implemented in custom hooks:

- **API Communication**: All API logic including HTTP requests, WebSocket communication, and associated types must be encapsulated in hooks, never directly in components.
- **Business Logic**: Encapsulate game rules and state management in hooks.
- **Reusability**: Design hooks to be reusable across different components.
- **Self-Contained**: Hooks should be self-contained with all required types and logic defined within the hook file.

Components should never directly communicate with the backend - they must only interact through hooks. This ensures:

1. UI components remain focused on presentation
2. API logic can be tested independently
3. Components can be easily reused with different data sources
4. Business logic is cleanly separated from data fetching

---

## API Endpoints and Communication

The frontend communicates with the backend through these endpoints:

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

## Project Structure

```
frontend/
├── src/
│   ├── components/         # UI components
│   ├── hooks/              # Custom hooks for business logic and API communication
│   ├── App.tsx             # Main application component
│   └── main.tsx            # Application entry point
```

Note:

- All API communication logic should be fully defined within hook files
- All types required for API communication should be defined in the hook files themselves
- Components should never communicate with the backend directly, only through hooks

---

## Development Workflow

1. **Hook Development**:

   - First, implement API communication hooks that encapsulate all backend interactions
   - Build higher-level business logic hooks that use the API hooks
   - Ensure hooks are testable and handle all edge cases (loading, errors, etc.)

2. **Component Development**:

   - Create generic, reusable UI components
   - Components should never contain API logic - they only use hooks
   - Pass all required data and callbacks via props

3. **Integration**:

   - Connect components to hooks in container components
   - Keep leaf components as pure UI elements
   - Use custom hooks to handle all side effects

4. **Testing**:
   - Write tests colocated with the files they're testing
   - Test components in isolation with mock hooks
   - Test hooks with mock API services
   - Verify both component rendering and hook behavior independently

---

## Technical Decisions

- **No Global State**: Improves testability and reduces coupling
- **Hooks for All Logic**: Completely separates UI from business logic and API communication
- **Generic Components**: Increases reusability and maintainability
- **Prop Drilling**: Makes data flow explicit and traceable
- **API Logic in Hooks Only**: Ensures components remain pure and testable without backend dependencies

---

## Implementation Guidelines

- **Functional Components Only**: No classes should be used; everything must be defined functionally with the syntax:

  ```tsx
  const ComponentName = (): ReactElement => {
    // Component implementation
    return (
      // JSX here
    );
  };
  ```

- **TypeScript with Strict Typing**: Only TypeScript shall be used with strict typing. The `any` type is not allowed under any circumstances.

- **Type over Interface**: Always use `type` instead of `interface` for defining TypeScript types. This provides better consistency and avoids potential issues with declaration merging.

- **Component Structure**: Each component should follow this structure:

  ```
  ComponentName/
    ComponentName.tsx        // Component logic and JSX
    ComponentName.test.tsx   // Component tests
    styles.module.css        // All styles for the component
  ```

- **CSS Modules**: All styles must be defined in the CSS file, using CSS modules for scoping. No inline styles or styled-components should be used.
