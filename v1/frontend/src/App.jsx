import { useState } from 'react';
import { MemoryRouter as Router, Route, Routes, Link } from 'react-router-dom';
import './App.css';
import { NavDashboard, NavSignIn } from './Navigation';
import Dashboard from './screens/Dashboard';
import SignIn from './screens/SignIn';

function App() {
    return (
        <Router>
            <Routes>
                <Route path={NavSignIn} element={<SignIn />} />
                <Route path={NavDashboard} element={<Dashboard />} />
            </Routes>
        </Router>
    );
}

export default App;
