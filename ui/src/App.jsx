import { Routes, Route } from "react-router-dom";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import IndexPage from "./pages/Index.jsx";
import SettingsAgentToken from "./pages/SettingsAgentToken.jsx";
import AdminUsers from "./pages/AdminUsers.jsx";
import Sessions from "./pages/Sessions.jsx";
import SessionDetail from "./pages/SessionDetail.jsx";
import LocalInstallGuide from "./pages/LocalInstallGuide.jsx";

export default function App() {
  return (
      <Routes>
          <Route path="/" element={<IndexPage />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/settings/token" element={<SettingsAgentToken />} />
          <Route path="/admin/users" element={<AdminUsers />} />
          <Route path="/sessions" element={<Sessions />} />
          <Route path="/sessions/:id" element={<SessionDetail />} />
          <Route path="/guide/local-install" element={<LocalInstallGuide />} />
      </Routes>
  );
}
