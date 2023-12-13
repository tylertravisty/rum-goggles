import { useEffect, useState } from 'react';
import { MemoryRouter as Router, Route, Routes, Link } from 'react-router-dom';

import './App.css';

import { NavSignIn, NavDashboard } from './components/Navigation';
import Dashboard from './components/Dashboard';
import SignIn from './components/SignIn';

function App() {
    return (
        <Router>
            <Routes>
                <Route path={NavSignIn} element={<SignIn />}></Route>
                <Route path={NavDashboard} element={<Dashboard />}></Route>
            </Routes>
        </Router>
    );
}

export default App;
