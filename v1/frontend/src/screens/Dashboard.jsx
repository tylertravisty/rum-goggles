import { useState } from 'react';

import { CircleGreenBackground, Heart } from '../assets';
import PageDetails from '../components/PageDetails';
import PageSideBar from '../components/PageSideBar';
import './Dashboard.css';

function Dashboard() {
    return (
        <div className='dashboard'>
            <PageSideBar />
            <PageDetails />
            <div style={{ backgroundColor: '#344453', width: '100%', height: '100%' }}></div>
        </div>
    );
}

export default Dashboard;
