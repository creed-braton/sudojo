const API_ENDPOINT: string = import.meta.env.VITE_API_ENDPOINT;
const INSECURE: boolean = import.meta.env.VITE_INSECURE === "true";
export const HTTP_URL: string = `${INSECURE ? "http" : "https"}://${API_ENDPOINT}`;
export const WS_URL: string = `${INSECURE ? "ws" : "wss"}://${API_ENDPOINT}`;
