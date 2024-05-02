import { useState } from 'react';

import { CircleGreenBackground, Heart } from '../assets';
import PageDetails from '../components/PageDetails';
import PageSideBar from '../components/PageSideBar';
import './Dashboard.css';
import ChatBot from '../components/ChatBot';

function Dashboard() {
    return (
        <div className='dashboard'>
            <PageSideBar />
            <PageDetails />
            <ChatBot />
        </div>
    );
}

export default Dashboard;
