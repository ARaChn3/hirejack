export type Payload = {
  id: string;
  name: string;
  description: string;
  impact: string;
  stealth: number[];
  locations: string[];
};

export type Category = {
  id: string;
  name: string;
  description: string;
  color: string;
  payloads: Payload[];
};

export type Status = "idle" | "armed" | "injecting" | "done" | "error";

export type FileMeta = {
  file: File;
  size: number;
  hash: string;
  pages: number | null;
};

export type ArtifactMeta = {
  blob: Blob;
  size: number;
  filename: string;
};

export type State = {
  file: FileMeta | null;
  categories: Category[];
  categoriesError: string | null;
  selected: Set<string>;
  status: Status;
  logs: string[];
  resultBlob: ArtifactMeta | null;
  error: string | null;
};

export type Action =
  | { type: "SET_CATEGORIES"; categories: Category[] }
  | { type: "SET_CATEGORIES_ERROR"; message: string }
  | { type: "SET_FILE"; file: FileMeta }
  | { type: "CLEAR_FILE" }
  | { type: "TOGGLE_PAYLOAD"; key: string }
  | { type: "SET_KEYS"; keys: string[]; on: boolean }
  | { type: "START_INJECT" }
  | { type: "LOG"; line: string }
  | { type: "INJECT_SUCCESS"; artifact: ArtifactMeta }
  | { type: "INJECT_ERROR"; message: string }
  | { type: "RUN_AGAIN" }
  | { type: "RESET" };
