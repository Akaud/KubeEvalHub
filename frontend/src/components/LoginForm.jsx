import React, { useState } from "react";
import { api } from "../lib/api";
import { useAuth } from "../auth/AuthContext";

export default function LoginForm() {
  const { setAuth } = useAuth();
  const [login, setLogin] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState(null);

  async function onSubmit(e) {
    e.preventDefault();
    setErr(null);

    try {
      const res = await api("/api/auth/login", {
        method: "POST",
        body: { login, password },
      });

      // backend returns { access_token, user: { id, username, email, ... } }
      setAuth(res.access_token, {
        id: res.user.id,
        username: res.user.username,
        email: res.user.email,
      });
    } catch (e2) {
      setErr(e2.message || "error");
    }
  }

  return (
    <form onSubmit={onSubmit} style={{ maxWidth: 420 }}>
      <h2>Login</h2>
      {err ? <div style={{ color: "crimson", marginBottom: 8 }}>{err}</div> : null}

      <label style={{ display: "block", marginTop: 8 }}>Email or username</label>
      <input
        value={login}
        onChange={(e) => setLogin(e.target.value)}
        style={{ width: "100%" }}
      />

      <label style={{ display: "block", marginTop: 8 }}>Password</label>
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        style={{ width: "100%" }}
      />

      <button type="submit" style={{ marginTop: 12 }}>
        Sign in
      </button>
    </form>
  );
}