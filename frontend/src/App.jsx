
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
        const data = await registerUser({
          name,
          email,
          password,
        });

        setMessage(data.message || "Registration successful!");
        setActivePage("login");
        setPassword("");
      } else {
        const data = await loginUser({
          email,
          password,
        });

        if (!data.token) {
          throw new Error("Login response lo token ledu.");
        }

        localStorage.setItem("token", data.token);

        setIsLoggedIn(true);
        setMessage("Login successful!");
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

  // Dashboard
  if (isLoggedIn) {
    return (
      <div className="app">
        <header className="navbar">
          <h2>GoCommerce</h2>

          <button type="button" onClick={handleLogout}>
            Logout
          </button>
        </header>

        <main className="container">
          <h1>Dashboard</h1>

          <p className="subtitle">
            Welcome to GoCommerce!
          </p>

          <section className="form-card">
            <h2>You are logged in</h2>

            <p>
              Your account login was successful.
            </p>

            <p>
              Order management features will be available here.
            </p>
          </section>
        </main>
      </div>
    );
  }

  // Login and Register
  return (
    <div className="app">
      <header className="navbar">
        <h2>GoCommerce</h2>
        <span>Microservices Platform</span>
      </header>

      <main className="container">
        <h1>Welcome to GoCommerce</h1>

        <p className="subtitle">
          Manage your account and orders in one place.
        </p>

        <div className="tabs">
          <button
            type="button"
            className={activePage === "login" ? "active" : ""}
            onClick={() => {
              setActivePage("login");
              setMessage("");
            }}
          >
            Login
          </button>

          <button
            type="button"
            className={activePage === "register" ? "active" : ""}
            onClick={() => {
              setActivePage("register");
              setMessage("");
            }}
          >
            Register
          </button>
        </div>

        <section className="form-card">
          <h2>
            {activePage === "login"
              ? "Login to your account"
              : "Create an account"}
          </h2>

          <form onSubmit={handleSubmit}>
            {activePage === "register" && (
              <>
                <label htmlFor="name">Name</label>

                <input
                  id="name"
                  type="text"
                  value={name}
                  onChange={(event) => setName(event.target.value)}
                  placeholder="Enter your name"
                  autoComplete="name"
                  required
                />
              </>
            )}

            <label htmlFor="email">Email</label>

            <input
              id="email"
              type="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="Enter your email"
              autoComplete="email"
              required
            />

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

            <button
              className="submit-btn"
              type="submit"
              disabled={loading}
            >
              {loading
                ? "Please wait..."
                : activePage === "login"
                  ? "Login"
                  : "Register"}
            </button>
          </form>

          {message && (
            <p role="status" aria-live="polite">
              {message}
            </p>
          )}
        </section>
      </main>
    </div>
  );
}

export default App;
