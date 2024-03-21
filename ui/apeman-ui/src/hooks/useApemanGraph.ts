import { useContext } from "react";
import { ApemanGraphContext } from "../components/ApemanGraphContext";

export function useApemanGraph() {
  const context = useContext(ApemanGraphContext);
  if (context === undefined) {
    throw new Error(
      "useApemanGraph must be used within an ApemanGraphProvider"
    );
  }
  return context;
}
