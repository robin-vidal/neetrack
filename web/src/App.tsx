import { AuthGate } from "./components/auth/AuthGate";

function App() {
  return (
    <AuthGate>
      {({ user, logout }) => (
        <main>
          <h1>Neetrack</h1>
          <p>
            Connecté en tant que {user.name}. {logout}
          </p>
        </main>
      )}
    </AuthGate>
  );
}

export default App;
