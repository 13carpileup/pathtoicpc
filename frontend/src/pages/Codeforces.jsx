import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import {
  authHeaders,
  clearSession,
  getStoredAuthToken,
  getStoredUser,
  jsonHeaders,
  storeUser
} from "../auth.js";

const challengeKey = "pathtoicpc.cfChallenge";

export default function Codeforces() {
  const [token, setToken] = useState(getStoredAuthToken);
  const [user, setUser] = useState(() => (getStoredAuthToken() ? getStoredUser() : null));
  const [handle, setHandle] = useState("");
  const [challenge, setChallenge] = useState(() => getStoredChallenge(getStoredUser()?.id));
  const [status, setStatus] = useState("");
  const [statusTone, setStatusTone] = useState("neutral");
  const [isStarting, setIsStarting] = useState(false);
  const [isVerifying, setIsVerifying] = useState(false);
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    if (!token) {
      setUser(null);
      setChallenge(null);
      return;
    }

    let ignore = false;

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
          setChallenge(getStoredChallenge(data.id));
        }
      } catch {
        if (!ignore) {
          clearSession(setToken, setUser);
          setChallenge(null);
          setStatusTone("error");
          setStatus("Please log in again.");
        }
      }
    }

    loadAccount();

    return () => {
      ignore = true;
    };
  }, [token]);

  useEffect(() => {
    const timer = window.setInterval(() => {
      setNow(Date.now());
    }, 1000);

    return () => {
      window.clearInterval(timer);
    };
  }, []);

  const remainingMs = challenge
    ? Math.max(0, new Date(challenge.expiresAt).getTime() - now)
    : 0;
  const isExpired = Boolean(challenge) && remainingMs <= 0;
  const problemURL = useMemo(() => getCodeforcesProblemURL(challenge?.problem), [challenge]);

  async function startChallenge(event) {
    event.preventDefault();

    const trimmedHandle = handle.trim();
    if (!trimmedHandle) {
      setStatusTone("error");
      setStatus("Enter a Codeforces handle.");
      return;
    }

    setIsStarting(true);
    setStatus("");

    try {
      const data = await apiRequest("/api/connect_cf", {
        method: "POST",
        headers: jsonHeaders(token),
        body: JSON.stringify({ codeforces_username: trimmedHandle })
      });

      const nextChallenge = {
        userId: user.id,
        handle: trimmedHandle,
        problem: data.problem,
        expiresAt: data.expiresAt
      };

      localStorage.setItem(challengeKey, JSON.stringify(nextChallenge));
      setChallenge(nextChallenge);
      setStatusTone("success");
      setStatus("Challenge created.");
    } catch (error) {
      setStatusTone("error");
      setStatus(error.message);
    } finally {
      setIsStarting(false);
    }
  }

  async function verifyChallenge() {
    setIsVerifying(true);
    setStatus("");

    try {
      const data = await apiRequest("/api/verify_cf", {
        method: "POST",
        headers: authHeaders(token)
      });

      localStorage.removeItem(challengeKey);
      setChallenge(null);
      setStatusTone("success");
      setStatus(data.message || "Codeforces account linked.");
    } catch (error) {
      setStatusTone("error");
      setStatus(error.message);
    } finally {
      setIsVerifying(false);
    }
  }

  return (
    <section className="page-section cf-page">
      <p className="eyebrow">Codeforces</p>
      <h1>Link your Codeforces handle.</h1>

      {status ? <p className={`form-status ${statusTone}`}>{status}</p> : null}

      {!token || !user ? (
        <div className="cf-panel cf-empty">
          <h2>Sign in first</h2>
          <p>Use an account before starting a Codeforces challenge.</p>
          <Link className="button-link" to="/account">
            Account
          </Link>
        </div>
      ) : (
        <div className="cf-workflow">
          <form className="cf-panel" onSubmit={startChallenge}>
            <div className="cf-panel-heading">
              <span>Handle</span>
              <strong>{user.username}</strong>
            </div>
            <label>
              Codeforces handle
              <input
                autoComplete="username"
                value={handle}
                onChange={(event) => setHandle(event.target.value)}
                placeholder="tourist"
                required
              />
            </label>
            <button type="submit" disabled={isStarting || isVerifying}>
              {isStarting ? "Generating..." : "Generate challenge"}
            </button>
          </form>

          <section className="cf-panel cf-challenge" aria-live="polite">
            <div className="cf-panel-heading">
              <span>Assigned problem</span>
              <strong>{challenge ? challenge.problem : "None"}</strong>
            </div>

            {challenge ? (
              <>
                <dl className="cf-meta">
                  <div>
                    <dt>Handle</dt>
                    <dd>{challenge.handle}</dd>
                  </div>
                  <div>
                    <dt>Time left</dt>
                    <dd className={isExpired ? "expired" : undefined}>
                      {isExpired ? "Expired" : formatDuration(remainingMs)}
                    </dd>
                  </div>
                  <div>
                    <dt>Expires</dt>
                    <dd>{new Date(challenge.expiresAt).toLocaleTimeString()}</dd>
                  </div>
                </dl>

                <div className="cf-actions">
                  {problemURL ? (
                    <a
                      className="button-link secondary"
                      href={problemURL}
                      rel="noreferrer"
                      target="_blank"
                    >
                      Open problem
                    </a>
                  ) : null}
                  <button
                    type="button"
                    onClick={verifyChallenge}
                    disabled={isVerifying || isExpired}
                  >
                    {isVerifying ? "Checking..." : "Verify submission"}
                  </button>
                </div>
              </>
            ) : (
              <p className="cf-muted">Generate a challenge to receive a problem.</p>
            )}
          </section>
        </div>
      )}
    </section>
  );
}

async function apiRequest(path, options) {
  const response = await fetch(path, options);
  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    throw new Error(data.error || "Request failed.");
  }

  return data;
}

function getStoredChallenge(userId) {
  if (!userId) {
    return null;
  }

  try {
    const challenge = JSON.parse(localStorage.getItem(challengeKey));
    if (!challenge || challenge.userId !== userId) {
      return null;
    }

    if (new Date(challenge.expiresAt).getTime() <= Date.now()) {
      localStorage.removeItem(challengeKey);
      return null;
    }

    return challenge;
  } catch {
    localStorage.removeItem(challengeKey);
    return null;
  }
}

function getCodeforcesProblemURL(problemId) {
  if (!problemId) {
    return "";
  }

  const match = /^(\d+)([A-Za-z]\d*)$/.exec(problemId);
  if (!match) {
    return "";
  }

  return `https://codeforces.com/problemset/problem/${match[1]}/${match[2]}`;
}

function formatDuration(milliseconds) {
  const totalSeconds = Math.ceil(milliseconds / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;

  return `${minutes}:${String(seconds).padStart(2, "0")}`;
}
