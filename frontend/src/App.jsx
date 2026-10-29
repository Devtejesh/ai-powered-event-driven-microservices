
import { useState } from "react";
import { registerUser, loginUser } from "./api";
import "./App.css";

function App() {
  const [activePage, setActivePage] = useState("login");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  const [isLoggedIn, setIsLoggedIn] = useState(
    Boolean(localStorage.getItem("token"))
  );

  async function handleSubmit(event) {
    event.preventDefault();
    setMessage("");
    setLoading(true);

    try {
      if (activePage === "register") {
        const data = await registerUser({ name, email, password });

        setMessage(data.message || "Registration successful!");
        setActivePage("login");
        setPassword("");
      } else {
        const data = await loginUser({ email, password });

        if (!data.token) {
          throw new Error("Login response lo token ledu.");
        }

        localStorage.setItem("token", data.token);
        setIsLoggedIn(true);
        setMessage("");
        setPassword("");
      }
    } catch (error) {
      setMessage(error.message || "Something went wrong.");
    } finally {
      setLoading(false);
    }
  }

  function handleLogout() {
    localStorage.removeItem("token");
    setIsLoggedIn(false);
    setActivePage("login");
    setName("");
    setEmail("");
    setPassword("");
    setMessage("");
  }

  if (isLoggedIn) {
    return (
      <div className="app-shell">
        <aside className="sidebar">
          <a className="brand" href="#dashboard">
            <span className="brand-mark">G</span>
            <span>Go<span className="brand-accent">Commerce</span></span>
          </a>

          <p className="sidebar-label">WORKSPACE</p>

          <a className="nav-item nav-active" href="#dashboard">
            <span className="nav-icon">▦</span>
            Dashboard
          </a>

          <div className="nav-item nav-muted">
            <span className="nav-icon">▤</span>
            Orders
            <span className="nav-soon">API next</span>
          </div>

          <div className="nav-item nav-muted">
            <span className="nav-icon">♙</span>
            Customers
          </div>

          <div className="nav-item nav-muted">
            <span className="nav-icon">⚙</span>
            Settings
          </div>

          <div className="sidebar-bottom">
            <div className="status-dot" />
            <div>
              <strong>Platform status</strong>
              <p>Dashboard connected</p>
            </div>
          </div>
        </aside>

        <main className="dashboard-main" id="dashboard">
          <header className="topbar">
            <div className="breadcrumb">
              Workspace <span>/</span> <strong>Dashboard</strong>
            </div>

            <div className="topbar-right">
              <span className="environment-tag">MICROSERVICES</span>
              <div className="avatar">
                {email ? email.charAt(0).toUpperCase() : "U"}
              </div>
              <button
                type="button"
                className="logout-btn"
                onClick={handleLogout}
              >
                Log out ↗
              </button>
            </div>
          </header>

          <section className="dashboard-content">
            <div className="welcome-row">
              <div>
                <p className="eyebrow">OVERVIEW</p>
                <h1>Welcome back<span className="heading-dot">.</span></h1>
                <p className="page-description">
                  Your GoCommerce workspace, all in one place.
                </p>
              </div>

              <div className="date-chip">
                <span>◷</span> Live workspace
              </div>
            </div>

            <section className="hero-panel">
              <div className="hero-copy">
                <span className="hero-tag">
                  <span className="tiny-dot" /> YOUR WORKSPACE
                </span>
                <h2>Commerce, connected.</h2>
                <p>
                  Manage your commerce platform through a unified
                  workspace built for your microservices architecture.
                </p>
                <div className="hero-pills">
                  <span>Go backend</span>
                  <span>REST APIs</span>
                  <span>Event-driven</span>
                </div>
              </div>

              <div className="hero-art" aria-hidden="true">
                <div className="orbit orbit-one" />
                <div className="orbit orbit-two" />
                <div className="orbit orbit-three" />
                <div className="core-node">G</div>
                <div className="orbit-node node-one">API</div>
                <div className="orbit-node node-two">DB</div>
                <div className="orbit-node node-three">↗</div>
              </div>
            </section>

            <div className="section-heading">
              <div>
                <h2>Workspace overview</h2>
                <p>Your main platform areas at a glance.</p>
              </div>
              <span className="demo-note">Live data not connected yet</span>
            </div>

            <section className="overview-grid">
              <article className="overview-card">
                <div className="card-top">
                  <span className="metric-icon purple-icon">▤</span>
                  <span className="card-arrow">↗</span>
                </div>
                <p className="metric-label">Order management</p>
                <h3>Orders</h3>
                <p className="metric-caption">
                  Create, view and manage orders
                </p>
                <div className="card-footer">
                  <span className="mini-line purple-line" />
                  <span>Order service</span>
                </div>
              </article>

              <article className="overview-card">
                <div className="card-top">
                  <span className="metric-icon blue-icon">♙</span>
                  <span className="card-arrow">↗</span>
                </div>
                <p className="metric-label">Account management</p>
                <h3>Customers</h3>
                <p className="metric-caption">
                  Account and user operations
                </p>
                <div className="card-footer">
                  <span className="mini-line blue-line" />
                  <span>User service</span>
                </div>
              </article>

              <article className="overview-card">
                <div className="card-top">
                  <span className="metric-icon green-icon">⌁</span>
                  <span className="card-arrow">↗</span>
                </div>
                <p className="metric-label">Event processing</p>
                <h3>Events</h3>
                <p className="metric-caption">
                  Event-driven service architecture
                </p>
                <div className="card-footer">
                  <span className="mini-line green-line" />
                  <span>Kafka integration</span>
                </div>
              </article>
            </section>

            <section className="bottom-grid">
              <article className="panel-card services-panel">
                <div className="panel-heading">
                  <div>
                    <h2>Platform services</h2>
                    <p>Services in your architecture</p>
                  </div>
                  <span className="panel-symbol">⌘</span>
                </div>

                <div className="service-row">
                  <span className="service-icon purple-icon">U</span>
                  <div className="service-details">
                    <strong>User service</strong>
                    <span>Authentication and accounts</span>
                  </div>
                  <span className="service-label">REST API</span>
                </div>

                <div className="service-row">
                  <span className="service-icon blue-icon">O</span>
                  <div className="service-details">
                    <strong>Order service</strong>
                    <span>Order lifecycle management</span>
                  </div>
                  <span className="service-label">REST API</span>
                </div>

                <div className="service-row">
                  <span className="service-icon green-icon">N</span>
                  <div className="service-details">
                    <strong>Notification service</strong>
                    <span>Event-based notifications</span>
                  </div>
                  <span className="service-label">Kafka</span>
                </div>
              </article>

              <article className="panel-card account-panel">
                <div className="panel-heading">
                  <div>
                    <h2>Your account</h2>
                    <p>Current session details</p>
                  </div>
                  <span className="account-avatar">
                    {email ? email.charAt(0).toUpperCase() : "U"}
                  </span>
                </div>

                <div className="account-field">
                  <span>Account name</span>
                  <strong>{name || "GoCommerce user"}</strong>
                </div>

                <div className="account-field">
                  <span>Email address</span>
                  <strong className="account-email">
                    {email || "Signed-in account"}
                  </strong>
                </div>

                <div className="session-status">
                  <span className="tiny-dot" />
                  Authenticated session
                </div>

                <button
                  type="button"
                  className="account-logout"
                  onClick={handleLogout}
                >
                  Sign out of account <span>→</span>
                </button>
              </article>
            </section>

            <footer className="dashboard-footer">
              <span>GoCommerce <span className="footer-separator">/</span> Workspace</span>
              <span>Built for connected commerce systems <span className="footer-heart">✦</span></span>
            </footer>
          </section>
        </main>
      </div>
    );
  }

  return (
    <div className="auth-shell">
      <section className="auth-showcase">
        <a className="brand auth-brand" href="#home">
          <span className="brand-mark">G</span>
          <span>Go<span className="brand-accent">Commerce</span></span>
        </a>

        <div className="showcase-content">
          <span className="hero-tag">
            <span className="tiny-dot" /> COMMERCE INFRASTRUCTURE
          </span>
          <h1>
            Everything<br />
            connected<span className="heading-dot">.</span>
          </h1>
          <p>
            A unified workspace for managing your commerce
            platform, accounts and order operations.
          </p>

          <div className="showcase-visual" aria-hidden="true">
            <div className="showcase-ring ring-a" />
            <div className="showcase-ring ring-b" />
            <div className="showcase-ring ring-c" />
            <div className="showcase-core">
              <span>G</span>
              <small>COMMERCE</small>
            </div>
            <div className="showcase-bubble bubble-a">API</div>
            <div className="showcase-bubble bubble-b">DB</div>
            <div className="showcase-bubble bubble-c">↗</div>
          </div>
        </div>

        <div className="showcase-footer">
          <span>GO MICROSERVICES PLATFORM</span>
          <span>SECURE · CONNECTED · SCALABLE</span>
        </div>
      </section>

      <section className="auth-side" id="home">
        <div className="auth-form-wrap">
          <div className="mobile-brand brand">
            <span className="brand-mark">G</span>
            <span>Go<span className="brand-accent">Commerce</span></span>
          </div>

          <div className="auth-heading">
            <span className="eyebrow">
              {activePage === "login" ? "WELCOME BACK" : "GET STARTED"}
            </span>
            <h2>
              {activePage === "login"
                ? "Sign in to your account"
                : "Create your account"}
            </h2>
            <p>
              {activePage === "login"
                ? "Enter your details to access your workspace."
                : "Set up your account to get started."}
            </p>
          </div>

          <div className="auth-tabs">
            <button
              type="button"
              className={activePage === "login" ? "auth-tab active-tab" : "auth-tab"}
              onClick={() => {
                setActivePage("login");
                setMessage("");
              }}
            >
              Sign in
            </button>
            <button
              type="button"
              className={activePage === "register" ? "auth-tab active-tab" : "auth-tab"}
              onClick={() => {
                setActivePage("register");
                setMessage("");
              }}
            >
              Create account
            </button>
          </div>

          <form className="auth-form" onSubmit={handleSubmit}>
            {activePage === "register" && (
              <div className="field-group">
                <label htmlFor="name">Full name</label>
                <input
                  id="name"
                  type="text"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="Your full name"
                  autoComplete="name"
                  required
                />
              </div>
            )}

            <div className="field-group">
              <label htmlFor="email">Email address</label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="you@example.com"
                autoComplete="email"
                required
              />
            </div>

            <div className="field-group">
              <label htmlFor="password">Password</label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="Enter your password"
                autoComplete={
                  activePage === "login"
                    ? "current-password"
                    : "new-password"
                }
                required
              />
            </div>

            {message && (
              <div className="form-message" role="status" aria-live="polite">
                {message}
              </div>
            )}

            <button
              className="auth-submit"
              type="submit"
              disabled={loading}
            >
              {loading
                ? "Please wait..."
                : activePage === "login"
                  ? "Sign in to workspace"
                  : "Create account"}
              {!loading && <span>→</span>}
            </button>
          </form>

          <div className="auth-security">
            <span className="security-icon">◇</span>
            <span>Your session is secured with token-based authentication.</span>
          </div>
        </div>

        <footer className="auth-bottom">
          <span>© GoCommerce</span>
          <span>Microservices workspace</span>
        </footer>
      </section>
    </div>
  );
}

export default App;
