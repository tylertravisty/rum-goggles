import { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { QueryAPI } from '../../wailsjs/go/main/App';

import './Dashboard.css';

function Dashboard() {
    const location = useLocation();
    const [streamKey, setStreamKey] = useState(location.state.streamKey);
    const [followers, setFollowers] = useState({});
    const [totalFollowers, setTotalFollowers] = useState('-');
    const [channelFollowers, setChannelFollowers] = useState('-');
    const [latestFollower, setLatestFollower] = useState('-');
    const [recentFollowers, setRecentFollowers] = useState([]);

    // useEffect(() => {
    //     QueryAPI(streamKey)
    //         .then((response) => {
    //             console.log(response);
    //             setFollowers(response);
    //             setChannelFollowers(response.num_followers);
    //             setTotalFollowers(response.num_followers_total);
    //             setLatestFollower(response.latest_follower.username);
    //             setRecentFollowers(response.recent_followers);
    //         })
    //         .catch((e) => console.log('Error:', e));
    // }, []);

    useEffect(() => {
        let interval = setInterval(() => {
            console.log('Query API');
            QueryAPI(streamKey)
                .then((response) => {
                    console.log(response);
                    setFollowers(response);
                    setChannelFollowers(response.num_followers);
                    setTotalFollowers(response.num_followers_total);
                    setLatestFollower(response.latest_follower.username);
                    setRecentFollowers(response.recent_followers);
                })
                .catch((e) => console.log('Error:', e));
        }, 10000);

        return () => {
            clearInterval(interval);
        };
    }, []);

    const dateDate = (date) => {
        const options = { month: 'short' };
        let month = new Intl.DateTimeFormat('en-US', options).format(date);
        let day = date.getDay();
        return month + ' ' + day;
    };

    const dateDay = (date) => {
        let now = new Date();
        let today = now.getDay();
        switch (date.getDay()) {
            case 0:
                return 'Sunday';
            case 1:
                return 'Monday';
            case 2:
                return 'Tuesday';
            case 3:
                return 'Wednesday';
            case 4:
                return 'Thursday';
            case 5:
                return 'Friday';
            case 6:
                return 'Saturday';
        }
    };

    const dateTime = (date) => {
        let now = new Date();
        let today = now.getDay();
        let day = date.getDay();

        if (today !== day) {
            return dateDay(date);
        }

        let hours24 = date.getHours();
        let hours = hours24 % 12 || 12;

        let minutes = date.getMinutes();

        let mer = 'pm';
        if (hours24 < 12) {
            mer = 'am';
        }

        return hours + ':' + minutes + ' ' + mer;
    };

    const dateString = (d) => {
        let now = new Date();
        let date = new Date(d);
        // Fix Rumble's timezone problem
        date.setHours(date.getHours() - 4);
        let diff = now - date;
        switch (true) {
            case diff < 60000:
                return 'Now';
            case diff < 3600000:
                let minutes = Math.floor(diff / 1000 / 60);
                let postfix = ' minutes ago';
                if (minutes == 1) {
                    postfix = ' minute ago';
                }
                return minutes + postfix;
            case diff < 86400000:
                return dateTime(date);
            case diff < 604800000:
                return dateDay(date);
            default:
                return dateDate(date);
        }
        console.log('Diff:', diff);
        return d;
    };

    return (
        <div id='Dashboard'>
            <span>Dashboard:</span>
            <div className='followers'>
                <div className='followers-header'>
                    <span className='followers-header-title'>Followers</span>
                    <div className='followers-header-highlights'>
                        <div className='followers-header-highlight'>
                            <span className='followers-header-highlight-count'>
                                {channelFollowers}
                            </span>
                            <span className='followers-header-highlight-description'>Channel</span>
                        </div>
                        <div className='followers-header-highlight'>
                            <span className='followers-header-highlight-count'>
                                {totalFollowers}
                            </span>
                            <span className='followers-header-highlight-description'>Total</span>
                        </div>
                    </div>
                </div>
                <div className='followers-list'>
                    {recentFollowers.map((follower, index) => (
                        <div className='followers-list-follower'>
                            <span className='followers-list-follower-username'>
                                {follower.username}
                            </span>
                            <span className='followers-list-follower-date'>
                                {dateString(follower.followed_on)}
                            </span>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
}

export default Dashboard;
