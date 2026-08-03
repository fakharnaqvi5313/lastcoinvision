import { RegistryClient } from "@lastcoinvision/sdk";

const REST_ENDPOINT = process.env.NEXT_PUBLIC_LCV_REST_URL ?? "http://localhost:1317";

export const registryClient = new RegistryClient(REST_ENDPOINT);
