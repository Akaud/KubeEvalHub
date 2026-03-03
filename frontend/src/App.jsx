import React, { useState } from "react";
import { AuthProvider, useAuth } from "./auth/AuthContext";
import LoginForm from "./components/LoginForm";
import RegisterForm from "./components/RegisterForm";
import Profile from "./components/Profile";

function Shell() {
  const { token } = useAuth();
  const [mode, setMode] = useState("login");

  if (token) {
    return (
      <div style={{ padding: 24 }}>
        <Profile />
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      <div style={{ marginBottom: 12 }}>
        <button onClick={() => setMode("login")}>Login</button>
        <button onClick={() => setMode("register")} style={{ marginLeft: 8 }}>
          Register
        </button>
      </div>

      {mode === "login" ? <LoginForm /> : <RegisterForm />}
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <Shell />
    </AuthProvider>
  );
}