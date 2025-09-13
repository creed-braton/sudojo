import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";

// Ensure cleanup after each test
afterEach(() => {
  cleanup();
});

// We want to use the real WebSocket implementation for integration tests
// No mocking is needed here
