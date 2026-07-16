import type { Action, State } from "@/types";

export const initialState: State = {
  file: null,
  categories: [],
  categoriesError: null,
  selected: new Set(),
  status: "idle",
  logs: [],
  resultBlob: null,
  error: null,
};

function deriveStatus(file: State["file"], selected: Set<string>): "idle" | "armed" {
  return file && selected.size > 0 ? "armed" : "idle";
}

export function reducer(state: State, action: Action): State {
  switch (action.type) {
    case "SET_CATEGORIES":
      return { ...state, categories: action.categories, categoriesError: null };

    case "SET_CATEGORIES_ERROR":
      return { ...state, categoriesError: action.message };

    case "SET_FILE": {
      const selected = new Set<string>();
      return {
        ...state,
        file: action.file,
        selected,
        status: "idle",
        logs: [],
        resultBlob: null,
        error: null,
      };
    }

    case "CLEAR_FILE":
      return {
        ...state,
        file: null,
        selected: new Set(),
        status: "idle",
        logs: [],
        resultBlob: null,
        error: null,
      };

    case "TOGGLE_PAYLOAD": {
      const selected = new Set(state.selected);
      if (selected.has(action.key)) selected.delete(action.key);
      else selected.add(action.key);
      return { ...state, selected, status: deriveStatus(state.file, selected) };
    }

    case "SET_KEYS": {
      const selected = new Set(state.selected);
      for (const key of action.keys) {
        if (action.on) selected.add(key);
        else selected.delete(key);
      }
      return { ...state, selected, status: deriveStatus(state.file, selected) };
    }

    case "START_INJECT":
      return { ...state, status: "injecting", logs: [], error: null, resultBlob: null };

    case "LOG":
      return { ...state, logs: [...state.logs, action.line] };

    case "INJECT_SUCCESS":
      return { ...state, status: "done", resultBlob: action.artifact };

    case "INJECT_ERROR":
      return { ...state, status: "error", error: action.message };

    case "RUN_AGAIN":
      return {
        ...state,
        status: deriveStatus(state.file, state.selected),
        logs: [],
        resultBlob: null,
        error: null,
      };

    case "RESET":
      return {
        ...initialState,
        categories: state.categories,
        categoriesError: state.categoriesError,
      };

    default:
      return state;
  }
}
