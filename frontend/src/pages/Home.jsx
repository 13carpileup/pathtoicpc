import { useEffect, useState } from "react";

export default function Home() {
  const [message, setMessage] = useState("Loading backend message...");

  useEffect(() => {
    let ignore = false;

    async function loadMessage() {
      try {
        const response = await fetch("/api/message");
        if (!response.ok) {
          throw new Error(`Request failed with ${response.status}`);
        }

        const data = await response.json();
        if (!ignore) {
          setMessage(data.message);
        }
      } catch {
        if (!ignore) {
          setMessage("Backend is not reachable yet.");
        }
      }
    }

    loadMessage();

    return () => {
      ignore = true;
    };
  }, []);

  return (
    <section className="page-section">
      <p className="eyebrow">Full-stack starter</p>
      <h1>Build your ICPC practice app from here.</h1>
      <p className="lead">
        This scaffold pairs a Go JSON API with a React frontend that already has
        client-side routing in place.
      </p>
      <div className="status-panel">
        <span>API message</span>
        <strong>{message}</strong>
      </div>
    </section>
  );
}
