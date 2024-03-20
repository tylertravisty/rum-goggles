import { CircleGreenBackground, Heart } from '../assets';
import ChannelSideBar from '../components/ChannelSideBar';
import './Dashboard.css';

function Dashboard() {
    return (
        <div className='dashboard'>
            <ChannelSideBar />
            <div style={{ backgroundColor: '#1f2e3c', width: '100%', height: '100%' }}></div>
        </div>
    );
}

export default Dashboard;
