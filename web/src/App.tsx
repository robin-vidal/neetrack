import { useState, type ReactNode } from "react";

import { AuthGate } from "./components/auth/AuthGate";
import { AdminDashboard } from "./components/admin/AdminDashboard";
import { Tracker } from "./components/tracker/Tracker";
import type { Me } from "./types";

interface AuthenticatedAppProps {
  user: Me;
  profileMenu: (onManageUsers: () => void) => ReactNode;
}

function AuthenticatedApp({ user, profileMenu }: AuthenticatedAppProps) {
  const [adminOpen, setAdminOpen] = useState(false);
  const openAdmin = () => setAdminOpen(true);

  if (adminOpen && user.is_admin) {
    return <AdminDashboard user={user} onReturn={() => setAdminOpen(false)} profileMenu={profileMenu(openAdmin)} />;
  }

  return <Tracker user={user} profileMenu={profileMenu} onOpenAdmin={openAdmin} />;
}

function App() {
  return <AuthGate>{({ user, profileMenu }) => <AuthenticatedApp user={user} profileMenu={profileMenu} />}</AuthGate>;
}

export default App;
