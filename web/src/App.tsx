import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { lazy, Suspense } from "react";
import Layout from "./components/Layout";
import Login from "./pages/Login";
import { ThemeProvider } from "./context/ThemeContext";
import { WebSocketProvider } from "./context/WebSocketContext";

const Dashboard = lazy(() => import("./pages/Dashboard"));
const Nodes = lazy(() => import("./pages/Nodes"));
const Services = lazy(() => import("./pages/Services"));
const Logs = lazy(() => import("./pages/Logs"));
const Rules = lazy(() => import("./pages/Rules"));
const Security = lazy(() => import("./pages/Security"));
const Configs = lazy(() => import("./pages/Configs"));
const Settings = lazy(() => import("./pages/Settings"));

function PageLoader() {
  return (
    <div className="flex items-center justify-center h-64">
      <div className="flex flex-col items-center gap-3">
        <div className="w-8 h-8 border-3 border-blue-600 border-t-transparent rounded-full animate-spin" />
        <span className="text-sm text-gray-500">加载中...</span>
      </div>
    </div>
  );
}

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const token = localStorage.getItem("token");
  if (!token) return <Navigate to="/login" />;
  return (
    <Layout>
      <Suspense fallback={<PageLoader />}>{children}</Suspense>
    </Layout>
  );
}

export default function App() {
  return (
    <ThemeProvider>
      <WebSocketProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/" element={<PrivateRoute><Dashboard /></PrivateRoute>} />
            <Route path="/nodes" element={<PrivateRoute><Nodes /></PrivateRoute>} />
            <Route path="/services" element={<PrivateRoute><Services /></PrivateRoute>} />
            <Route path="/logs" element={<PrivateRoute><Logs /></PrivateRoute>} />
            <Route path="/rules" element={<PrivateRoute><Rules /></PrivateRoute>} />
            <Route path="/security" element={<PrivateRoute><Security /></PrivateRoute>} />
            <Route path="/configs" element={<PrivateRoute><Configs /></PrivateRoute>} />
            <Route path="/settings" element={<PrivateRoute><Settings /></PrivateRoute>} />
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </BrowserRouter>
      </WebSocketProvider>
    </ThemeProvider>
  );
}
