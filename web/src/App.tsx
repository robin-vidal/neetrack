import { AuthGate } from "./components/auth/AuthGate";
import { Tracker } from "./components/tracker/Tracker";

function App() {
  return <AuthGate>{({ user, logout }) => <Tracker user={user} logout={logout} />}</AuthGate>;
}

export default App;
