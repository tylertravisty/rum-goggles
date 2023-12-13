import { useState } from 'react';
import { useLocation } from 'react-router-dom';

import './Dashboard.css';

function Dashboard() {
    const location = useLocation();
    const [streamKey, setStreamKey] = useState(location.state.streamKey);
    return (
        <div id='Dashboard'>
            <span>Dashboard: {streamKey}</span>
        </div>
    );
}

export default Dashboard;
