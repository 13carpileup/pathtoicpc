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

const challengeKey = "pathtoicpc.activeChallenge";
const linkingKey = "pathtoicpc.cfLinking";
const linkedAccountKey = "pathtoicpc.linkedCodeforces";

const difficulties = [
  {
    value: "EASY",
    label: "Easy",
    delta: "-200"
  },
  {
    value: "MEDIUM",
    label: "Medium",
    delta: "Current"
  },
  {
    value: "HARD",
    label: "Hard",
    delta: "+200"
  }
];

export default function Codeforces() {
  const [token, setToken] = useState(getStoredAuthToken);
  const [user, setUser] = useState(() => (getStoredAuthToken() ? getStoredUser() : null));
  const [handle, setHandle] = useState("");
  const [linking, setLinking] = useState(() => getStoredLinking(getStoredUser()?.id));
  const [linkedHandle, setLinkedHandle] = useState(() => getStoredLinkedHandle(getStoredUser()?.id));
  const [difficulty, setDifficulty] = useState("MEDIUM");
  const [challenge, setChallenge] = useState(() => getStoredChallenge(getStoredUser()?.id));
  const [status, setStatus] = useState("");
  const [statusTone, setStatusTone] = useState("neutral");
  const [isCreatingLink, setIsCreatingLink] = useState(false);
  const [isVerifyingLink, setIsVerifyingLink] = useState(false);
  const [isStarting, setIsStarting] = useState(false);
  const [isVerifying, setIsVerifying] = useState(false);
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    if (!token) {
      setUser(null);
      setLinking(null);
      setLinkedHandle("");
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
          setLinking(getStoredLinking(data.id));
          setLinkedHandle(getStoredLinkedHandle(data.id));
          setChallenge(getStoredChallenge(data.id));
        }
      } catch {
        if (!ignore) {
          clearSession(setToken, setUser);
          setLinking(null);
          setLinkedHandle("");
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

  const linkingExpiryTime = getLinkingExpiry(linking);
  const linkingRemainingMs = linkingExpiryTime
    ? Math.max(0, new Date(linkingExpiryTime).getTime() - now)
    : 0;
  const isLinkingExpired = Boolean(linking) && linkingRemainingMs <= 0;
  const linkingProblemURL = useMemo(
    () => getCodeforcesProblemURL(linking?.problem),
    [linking?.problem]
  );

  const expiryTime = getChallengeExpiry(challenge);
  const remainingMs = expiryTime ? Math.max(0, new Date(expiryTime).getTime() - now) : 0;
  const isExpired = Boolean(challenge) && remainingMs <= 0;
  const problemURL = useMemo(
    () => getCodeforcesProblemURL(challenge?.problem_id),
    [challenge?.problem_id]
  );

  async function startCodeforcesLink(event) {
    event.preventDefault();

    const trimmedHandle = handle.trim();
    if (!trimmedHandle) {
      setStatusTone("error");
      setStatus("Enter a Codeforces handle.");
      return;
    }

    setIsCreatingLink(true);
    setStatus("");

    try {
      const data = await apiRequest("/api/connect_cf", {
        method: "POST",
        headers: jsonHeaders(token),
        body: JSON.stringify({ codeforces_username: trimmedHandle })
      });

      const nextLinking = normalizeLinking(data, user.id, trimmedHandle);
      localStorage.setItem(linkingKey, JSON.stringify(nextLinking));
      setLinking(nextLinking);
      setStatusTone("success");
      setStatus("Codeforces link challenge created.");
    } catch (error) {
      setStatusTone("error");
      setStatus(error.message);
    } finally {
      setIsCreatingLink(false);
    }
  }

  async function verifyCodeforcesLink() {
    if (!linking) {
      setStatusTone("error");
      setStatus("Create a Codeforces link challenge first.");
      return;
    }

    setIsVerifyingLink(true);
    setStatus("");

    try {
      const data = await apiRequest("/api/verify_cf", {
        method: "POST",
        headers: authHeaders(token)
      });

      localStorage.removeItem(linkingKey);
      storeLinkedHandle(user.id, linking.handle);
      setLinkedHandle(linking.handle);
      setLinking(null);
      setStatusTone("success");
      setStatus(data.message || "Codeforces account linked.");
    } catch (error) {
      setStatusTone("error");
      setStatus(error.message);
    } finally {
      setIsVerifyingLink(false);
    }
  }

  async function startChallenge(event) {
    event.preventDefault();

    if (!token || !user) {
      setStatusTone("error");
      setStatus("Please log in first.");
      return;
    }

    setIsStarting(true);
    setStatus("");

    try {
      const data = await apiRequest("/api/chal", {
        method: "POST",
        headers: jsonHeaders(token),
        body: JSON.stringify({ challenge_type: difficulty })
      });

      const nextChallenge = normalizeChallenge(data, user.id, difficulty);
      localStorage.setItem(challengeKey, JSON.stringify(nextChallenge));
      setChallenge(nextChallenge);
      setStatusTone("success");
      setStatus("Challenge started.");
    } catch (error) {
      setStatusTone("error");
      setStatus(error.message);
    } finally {
      setIsStarting(false);
    }
  }

  async function verifyChallenge() {
    if (!challenge?.challenge_id) {
      setStatusTone("error");
      setStatus("Start a challenge first.");
      return;
    }

    setIsVerifying(true);
    setStatus("");

    try {
      const data = await apiRequest("/api/chal-update", {
        method: "POST",
        headers: jsonHeaders(token),
        body: JSON.stringify({ challenge_id: challenge.challenge_id })
      });

      localStorage.removeItem(challengeKey);
      setChallenge(null);
      setStatusTone("success");
      setStatus(data.message || "Challenge solved.");
    } catch (error) {
      setStatusTone("error");
      setStatus(error.message);
    } finally {
      setIsVerifying(false);
    }
  }

  return (
    <section className="page-section cf-page">
      <p className="eyebrow">Practice</p>
      <h1>Timed Codeforces challenge.</h1>

      {status ? <p className={`form-status ${statusTone}`}>{status}</p> : null}

      {!token || !user ? (
        <div className="cf-panel cf-empty">
          <h2>Sign in first</h2>
          <p>Use an account before starting a challenge.</p>
          <Link className="button-link" to="/account">
            Account
          </Link>
        </div>
      ) : (
        <>
          <section className="cf-panel cf-linking" aria-live="polite">
            <div className="cf-panel-heading">
              <span>Codeforces account</span>
              <strong>{linkedHandle || "Not linked in this browser"}</strong>
            </div>

            <form className="cf-link-form" onSubmit={startCodeforcesLink}>
              <label>
                Handle
                <input
                  autoComplete="username"
                  value={handle}
                  onChange={(event) => setHandle(event.target.value)}
                  placeholder="tourist"
                  required
                />
              </label>
              <button type="submit" disabled={isCreatingLink || isVerifyingLink}>
                {isCreatingLink ? "Creating..." : linking ? "New link challenge" : "Link account"}
              </button>
            </form>

            {linking ? (
              <>
                <dl className="cf-meta">
                  <div>
                    <dt>Handle</dt>
                    <dd>{linking.handle}</dd>
                  </div>
                  <div>
                    <dt>Problem</dt>
                    <dd>{linking.problem}</dd>
                  </div>
                  <div>
                    <dt>Time left</dt>
                    <dd className={isLinkingExpired ? "expired" : undefined}>
                      {isLinkingExpired ? "Expired" : formatDuration(linkingRemainingMs)}
                    </dd>
                  </div>
                </dl>

                <div className="cf-actions">
                  {linkingProblemURL ? (
                    <a
                      className="button-link secondary"
                      href={linkingProblemURL}
                      rel="noreferrer"
                      target="_blank"
                    >
                      Open link problem
                    </a>
                  ) : null}
                  <button
                    type="button"
                    onClick={verifyCodeforcesLink}
                    disabled={isCreatingLink || isVerifyingLink || isLinkingExpired}
                  >
                    {isVerifyingLink ? "Checking..." : "Verify account"}
                  </button>
                </div>
              </>
            ) : null}
          </section>

          <div className="cf-workflow">
            <form className="cf-panel cf-generator" onSubmit={startChallenge}>
              <div className="cf-panel-heading">
                <span>Player</span>
                <strong>{user.username}</strong>
              </div>

              <fieldset className="difficulty-picker">
                <legend>Difficulty</legend>
                <div>
                  {difficulties.map((option) => (
                    <label key={option.value}>
                      <input
                        checked={difficulty === option.value}
                        name="difficulty"
                        onChange={() => setDifficulty(option.value)}
                        type="radio"
                        value={option.value}
                      />
                      <span>
                        <strong>{option.label}</strong>
                        <small>{option.delta}</small>
                      </span>
                    </label>
                  ))}
                </div>
              </fieldset>

              <button type="submit" disabled={isStarting || isVerifying}>
                {isStarting ? "Starting..." : challenge ? "Start new challenge" : "Start challenge"}
              </button>
            </form>

            <section className="cf-panel cf-challenge" aria-live="polite">
              <div className="cf-panel-heading">
                <span>Assigned problem</span>
                <strong>{challenge ? challenge.problem_id : "None"}</strong>
              </div>

              {challenge ? (
                <>
                  <dl className="cf-meta">
                    <div>
                      <dt>Difficulty</dt>
                      <dd>{formatDifficulty(challenge.difficulty)}</dd>
                    </div>
                    <div>
                      <dt>Time left</dt>
                      <dd className={isExpired ? "expired" : undefined}>
                        {isExpired ? "Expired" : formatDuration(remainingMs)}
                      </dd>
                    </div>
                    <div>
                      <dt>Expires</dt>
                      <dd>{formatDateTime(expiryTime)}</dd>
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
                      disabled={isVerifying || isStarting || isExpired}
                    >
                      {isVerifying ? "Checking..." : "Verify solve"}
                    </button>
                  </div>
                </>
              ) : (
                <p className="cf-muted">No active challenge.</p>
              )}
            </section>
          </div>

          {challenge ? (
            <article className="problem-panel">
              <header>
                <span>Problem statement</span>
                <strong>{challenge.problem_id}</strong>
              </header>
              {challenge.challenge_text ? (
                <div
                  className="challenge-statement"
                  dangerouslySetInnerHTML={{ __html: challenge.challenge_text }}
                />
              ) : (
                <p className="cf-muted">Problem text is unavailable.</p>
              )}
            </article>
          ) : null}
        </>
      )}
    </section>
  );
}

async function apiRequest(path, options) {
  const response = await fetch(path, options);
  const data = await response.json().catch(() => ({}));

  if (data.error) {
    throw new Error(data.error);
  }

  if (!response.ok) {
    throw new Error(`Request failed with ${response.status}.`);
  }

  return data;
}

function normalizeLinking(data, userId, handle) {
  if (!data || !data.problem || !data.expiresAt) {
    throw new Error("Codeforces link response was incomplete.");
  }

  return {
    userId,
    handle,
    problem: data.problem,
    expiresAt: data.expiresAt
  };
}

function normalizeChallenge(data, userId, difficulty) {
  if (!data || !data.challenge_id || !data.problem_id || !data.expiry_time) {
    throw new Error("Challenge response was incomplete.");
  }

  return {
    userId,
    difficulty,
    challenge_id: data.challenge_id,
    user_id: data.user_id,
    problem_id: data.problem_id,
    solved: Boolean(data.solved),
    creation_time: data.creation_time,
    expiry_time: data.expiry_time,
    challenge_text: data.challenge_text || ""
  };
}

function getStoredLinking(userId) {
  if (!userId) {
    return null;
  }

  try {
    const linking = JSON.parse(localStorage.getItem(linkingKey));
    if (!linking || linking.userId !== userId) {
      return null;
    }

    const expiryTime = getLinkingExpiry(linking);
    if (!expiryTime || new Date(expiryTime).getTime() <= Date.now()) {
      localStorage.removeItem(linkingKey);
      return null;
    }

    return linking;
  } catch {
    localStorage.removeItem(linkingKey);
    return null;
  }
}

function getStoredLinkedHandle(userId) {
  if (!userId) {
    return "";
  }

  try {
    const linked = JSON.parse(localStorage.getItem(linkedAccountKey));
    if (!linked || linked.userId !== userId) {
      return "";
    }

    return linked.handle || "";
  } catch {
    localStorage.removeItem(linkedAccountKey);
    return "";
  }
}

function storeLinkedHandle(userId, handle) {
  localStorage.setItem(linkedAccountKey, JSON.stringify({ userId, handle }));
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

    const expiryTime = getChallengeExpiry(challenge);
    if (!expiryTime || new Date(expiryTime).getTime() <= Date.now()) {
      localStorage.removeItem(challengeKey);
      return null;
    }

    return challenge;
  } catch {
    localStorage.removeItem(challengeKey);
    return null;
  }
}

function getLinkingExpiry(linking) {
  return linking?.expiresAt || "";
}

function getChallengeExpiry(challenge) {
  return challenge?.expiry_time || challenge?.expiresAt || "";
}

function getCodeforcesProblemURL(problemId) {
  if (!problemId) {
    return "";
  }

  const match = /^(\d+)\/?([A-Za-z][A-Za-z0-9]*)$/.exec(problemId);
  if (!match) {
    return "";
  }

  return `https://codeforces.com/problemset/problem/${match[1]}/${match[2]}`;
}

function formatDifficulty(value) {
  const option = difficulties.find((difficulty) => difficulty.value === value);
  return option?.label || "Medium";
}

function formatDateTime(value) {
  if (!value) {
    return "Unknown";
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "Unknown";
  }

  return date.toLocaleString([], {
    dateStyle: "medium",
    timeStyle: "short"
  });
}

function formatDuration(milliseconds) {
  const totalSeconds = Math.max(0, Math.ceil(milliseconds / 1000));
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, "0")}:${String(seconds).padStart(2, "0")}`;
  }

  return `${minutes}:${String(seconds).padStart(2, "0")}`;
}
