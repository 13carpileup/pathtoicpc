const stats = [
  { label: "Routes", value: "4" },
  { label: "API endpoints", value: "8" },
  { label: "Stack", value: "Go + React" }
];

export default function Dashboard() {
  return (
    <section className="page-section">
      <p className="eyebrow">Dashboard</p>
      <h1>Project snapshot.</h1>
      <div className="stat-grid">
        {stats.map((stat) => (
          <article className="stat-card" key={stat.label}>
            <span>{stat.label}</span>
            <strong>{stat.value}</strong>
          </article>
        ))}
      </div>
    </section>
  );
}
