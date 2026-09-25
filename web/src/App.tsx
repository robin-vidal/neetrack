import { AuthGate } from "./components/auth/AuthGate";
import { Tracker } from "./components/tracker/Tracker";

function App() {
  return <AuthGate>{({ user, profileMenu }) => <Tracker user={user} profileMenu={profileMenu} />}</AuthGate>;
}

export default App;
