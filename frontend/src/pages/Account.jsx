import { useEffect, useState } from "react";
import {
  authHeaders,
  authRequest,
  clearSession,
  getStoredAuthToken,
  getStoredUser,
  saveSession,
  storeUser
} from "../auth.js";

const emptyLogin = {
  identifier: "",
  password: ""
};

const emptyRegistration = {
  email: "",
  username: "",
  password: ""
};

export default function Account() {
  const [token, setToken] = useState(getStoredAuthToken);
  const [user, setUser] = useState(() => (getStoredAuthToken() ? getStoredUser() : null));
  const [login, setLogin] = useState(emptyLogin);
  const [registration, setRegistration] = useState(emptyRegistration);
  const [status, setStatus] = useState("");
  const [isBusy, setIsBusy] = useState(false);
  const [isCheckingSession, setIsCheckingSession] = useState(
    () => Boolean(getStoredAuthToken()) && !getStoredUser()
  );

  useEffect(() => {
    if (!token) {
      setUser(null);
      setIsCheckingSession(false);
      return;
    }

    let ignore = false;
    setIsCheckingSession(!user);

    async function loadAccount() {
      try {
        const response = await fetch("/api/auth/me", {
          headers: authHeaders(token)
        });

        if (!response.ok) {
          throw new Error("Session expired.");
        }

        const data = await response.json();
        if (!ignore) {
          storeUser(data);
          setUser(data);
          setStatus("");
        }
      } catch {
        if (!ignore) {
          clearSession(setToken, setUser);
          setStatus("Please log in.");
        }
      } finally {
        if (!ignore) {
          setIsCheckingSession(false);
        }
      }
    }

    loadAccount();

    return () => {
      ignore = true;
    };
  }, [token]);

  async function submitRegistration(event) {
    event.preventDefault();
    setIsBusy(true);
    setStatus("");

    try {
      const data = await authRequest("/api/auth/register", registration);
      saveSession(data, setToken, setUser);
      setRegistration(emptyRegistration);
      setStatus("Account created.");
    } catch (error) {
      setStatus(error.message);
    } finally {
      setIsBusy(false);
    }
  }

  async function submitLogin(event) {
    event.preventDefault();
    setIsBusy(true);
    setStatus("");

    try {
      const data = await authRequest("/api/auth/login", login);
      saveSession(data, setToken, setUser);
      setLogin(emptyLogin);
      setStatus("Logged in.");
    } catch (error) {
      setStatus(error.message);
    } finally {
      setIsBusy(false);
    }
  }

  async function submitLogout() {
    if (!token) {
      clearSession(setToken, setUser);
      return;
    }

    setIsBusy(true);
    setStatus("");

    try {
      await fetch("/api/auth/logout", {
        method: "POST",
        headers: authHeaders(token)
      });
    } finally {
      clearSession(setToken, setUser);
      setStatus("Logged out.");
      setIsBusy(false);
    }
  }

  return (
    <section className="page-section account-page">
      <p className="eyebrow">Account</p>
      <h1>
        {user
          ? `Welcome, ${user.username}.`
          : isCheckingSession
            ? "Checking your session."
            : "Log in or create an account."}
      </h1>

      {status ? <p className="form-status">{status}</p> : null}

      {user ? (
        <div className="account-summary">
          <div>
            <span>Email</span>
            <strong>{user.email}</strong>
          </div>
          <div>
            <span>Joined</span>
            <strong>{new Date(user.createdAt).toLocaleDateString()}</strong>
          </div>
          <button type="button" onClick={submitLogout} disabled={isBusy}>
            Log out
          </button>
        </div>
      ) : isCheckingSession ? (
        <div className="account-summary session-loading">
          <div>
            <span>Session</span>
            <strong>Loading...</strong>
          </div>
        </div>
      ) : (
        <div className="auth-grid">
          <form className="auth-panel" onSubmit={submitLogin}>
            <h2>Log in</h2>
            <label>
              Email or username
              <input
                autoComplete="username"
                value={login.identifier}
                onChange={(event) =>
                  setLogin((current) => ({ ...current, identifier: event.target.value }))
                }
                required
              />
            </label>
            <label>
              Password
              <input
                autoComplete="current-password"
                type="password"
                value={login.password}
                onChange={(event) =>
                  setLogin((current) => ({ ...current, password: event.target.value }))
                }
                required
              />
            </label>
            <button type="submit" disabled={isBusy}>
              Log in
            </button>
          </form>

          <form className="auth-panel" onSubmit={submitRegistration}>
            <h2>Create account</h2>
            <label>
              Email
              <input
                autoComplete="email"
                type="email"
                value={registration.email}
                onChange={(event) =>
                  setRegistration((current) => ({ ...current, email: event.target.value }))
                }
                required
              />
            </label>
            <label>
              Username
              <input
                autoComplete="username"
                minLength={3}
                maxLength={64}
                pattern="[A-Za-z0-9_]+"
                value={registration.username}
                onChange={(event) =>
                  setRegistration((current) => ({ ...current, username: event.target.value }))
                }
                required
              />
            </label>
            <label>
              Password
              <input
                autoComplete="new-password"
                minLength={8}
                type="password"
                value={registration.password}
                onChange={(event) =>
                  setRegistration((current) => ({ ...current, password: event.target.value }))
                }
                required
              />
            </label>
            <button type="submit" disabled={isBusy}>
              Create account
            </button>
          </form>
        </div>
      )}
    </section>
  );
}
