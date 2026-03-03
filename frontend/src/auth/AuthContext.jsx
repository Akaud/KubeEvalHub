import React, { createContext, useContext, useMemo, useState } from "react";
import { clearAuth, loadAuth, saveAuth } from "../lib/storage.js";

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const initial = useMemo(() => loadAuth(), []);
  const [token, setToken] = useState(initial.token);
  const [user, setUser] = useState(initial.user);

  function setAuth(nextToken, nextUser) {
    setToken(nextToken);
    setUser(nextUser);
    saveAuth(nextToken, nextUser);
  }

  function logout() {
    setToken(null);
    setUser(null);
    clearAuth();
  }

  return (
    <AuthContext.Provider value={{ token, user, setAuth, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("AuthProvider missing");
  return ctx;
}