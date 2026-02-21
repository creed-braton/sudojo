const API_ENDPOINT: string | undefined = import.meta.env.VITE_API_ENDPOINT;
const INSECURE: boolean = import.meta.env.VITE_INSECURE === "true";

export const HTTP_URL: string = API_ENDPOINT
  ? `${INSECURE ? "http" : "https"}://${API_ENDPOINT}`
  : `${window.location.origin}/api`;

export const WS_URL: string = API_ENDPOINT
  ? `${INSECURE ? "ws" : "wss"}://${API_ENDPOINT}`
  : `${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}/api`;
