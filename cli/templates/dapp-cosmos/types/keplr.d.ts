import type { KeplrLike } from "@lastcoinvision/sdk";

declare global {
  interface Window {
    keplr?: KeplrLike;
  }
}

export {};
