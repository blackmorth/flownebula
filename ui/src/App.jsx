import { Routes, Route } from "react-router-dom";
import Login from "./pages/Login";
import Register from "./pages/Register";
import Dashboard from "./pages/Dashboard";
import IndexPage from "./pages/Index.jsx";
import Sessions from "./pages/Sessions.jsx";
import SessionDetail from "./pages/SessionDetail.jsx";
import Scripts from "./pages/Scripts.jsx";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<IndexPage />} />
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/dashboard" element={<Dashboard />} />
      <Route path="/scripts" element={<Scripts />} />
      <Route path="/sessions" element={<Sessions />} />
      <Route path="/sessions/:id" element={<SessionDetail />} />
    </Routes>
  );
}
