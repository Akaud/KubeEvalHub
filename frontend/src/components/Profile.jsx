import React from "react";
import { useAuth } from "../auth/AuthContext";

export default function Profile() {
  const { user, logout } = useAuth();

  return (
    <div style={{ maxWidth: 420 }}>
      <h2>Profile</h2>
      <div>Username: <b>{user ? user.username : ""}</b></div>

      <button onClick={logout} style={{ marginTop: 12 }}>
        Logout
      </button>
    </div>
  );
}