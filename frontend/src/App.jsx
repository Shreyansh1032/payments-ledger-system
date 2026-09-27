import { Routes, Route, Navigate } from 'react-router-dom';
import { Signin } from './pages/Signin';
import { Signup } from './pages/Signup';
import { Dashboard } from './pages/Dashboard';
import { SendMoney } from './pages/SendMoney';
import { Transactions } from './pages/Transactions';
import { Deposit } from './pages/Deposit';
import { Profile } from './pages/Profile';
import { Scan } from './pages/Scan';
import { Requests } from './pages/Requests';
import { Split } from './pages/Split';
import { Settings } from './pages/Settings';
import { Insights } from './pages/Insights';
import { ProtectedRoute } from './components/ProtectedRoute';

function App() {
  return (
    <Routes>
      <Route path="/" element={<Navigate to="/signin" replace />} />
      <Route path="/signin" element={<Signin />} />
      <Route path="/signup" element={<Signup />} />
      <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
      <Route path="/send" element={<ProtectedRoute><SendMoney /></ProtectedRoute>} />
      <Route path="/transactions" element={<ProtectedRoute><Transactions /></ProtectedRoute>} />
      <Route path="/deposit" element={<ProtectedRoute><Deposit /></ProtectedRoute>} />
      <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
      <Route path="/scan" element={<ProtectedRoute><Scan /></ProtectedRoute>} />
      <Route path="/requests" element={<ProtectedRoute><Requests /></ProtectedRoute>} />
      <Route path="/split" element={<ProtectedRoute><Split /></ProtectedRoute>} />
      <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
      <Route path="/insights" element={<ProtectedRoute><Insights /></ProtectedRoute>} />
    </Routes>
  );
}

export default App;
