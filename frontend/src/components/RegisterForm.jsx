import React, { useState } from "react";
import { api } from "../lib/api";
import { useAuth } from "../auth/AuthContext";

export default function RegisterForm() {
  const { setAuth } = useAuth();
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState(null);

  async function onSubmit(e) {
    e.preventDefault();
    setErr(null);

    try {
      const res = await api("/api/auth/register", {
        method: "POST",
        body: { username, email, password },
      });

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
      <h2>Register</h2>
      {err ? <div style={{ color: "crimson", marginBottom: 8 }}>{err}</div> : null}

      <label style={{ display: "block", marginTop: 8 }}>Username</label>
      <input
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        style={{ width: "100%" }}
      />

      <label style={{ display: "block", marginTop: 8 }}>Email</label>
      <input
        value={email}
        onChange={(e) => setEmail(e.target.value)}
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
        Create account
      </button>
    </form>
  );
}