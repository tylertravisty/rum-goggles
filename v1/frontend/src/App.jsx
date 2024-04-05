import { useState } from 'react';
import { MemoryRouter as Router, Route, Routes, Link } from 'react-router-dom';
import './App.css';
import { NavDashboard, NavSignIn, NavStartup } from './Navigation';
import Dashboard from './screens/Dashboard';
import SignIn from './screens/SignIn';
import Startup from './screens/Startup';

function App() {
    return (
        <Router>
            <Routes>
                <Route path={NavStartup} element={<Startup />} />
                <Route path={NavSignIn} element={<SignIn />} />
                <Route path={NavDashboard} element={<Dashboard />} />
            </Routes>
        </Router>
    );
}

export default App;
