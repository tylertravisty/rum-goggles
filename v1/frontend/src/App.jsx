import { useState } from 'react';
import { MemoryRouter as Router, Route, Routes, Link } from 'react-router-dom';
import './App.css';
import { NavSignIn } from './Navigation';
import SignIn from './screens/SignIn';

function App() {
    return (
        <Router>
            <Routes>
                <Route path={NavSignIn} element={<SignIn />} />
            </Routes>
        </Router>
    );
}

export default App;
