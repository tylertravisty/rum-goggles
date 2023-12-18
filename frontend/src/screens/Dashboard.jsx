import { useEffect, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { Start, Stop } from '../../wailsjs/go/api/Api';

import './Dashboard.css';
import { EventsEmit, EventsOn } from '../../wailsjs/runtime/runtime';
import { Heart, Star } from '../assets/icons';
import Highlight from '../components/Highlight';
import StreamEvent from '../components/StreamEvent';
import StreamActivity from '../components/StreamActivity';
import StreamChat from '../components/StreamChat';
import StreamInfo from '../components/StreamInfo';

function Dashboard() {
    const location = useLocation();
    const [refresh, setRefresh] = useState(false);
    const [active, setActive] = useState(false);
    const [streamKey, setStreamKey] = useState(location.state.streamKey);
    const [channelName, setChannelName] = useState('');
    const [followers, setFollowers] = useState({});
    const [totalFollowers, setTotalFollowers] = useState(0);
    const [channelFollowers, setChannelFollowers] = useState(0);
    const [recentFollowers, setRecentFollowers] = useState([]);
    const [subscribers, setSubscribers] = useState({});
    const [subscriberCount, setSubscriberCount] = useState(0);
    const [recentSubscribers, setRecentSubscribers] = useState([]);
    const [streamCategories, setStreamCategories] = useState({
        primary: { title: '' },
        secondary: { title: '' },
    });
    const [streamLikes, setStreamLikes] = useState(0);
    const [streamLive, setStreamLive] = useState(false);
    const [streamDislikes, setStreamDislikes] = useState(0);
    const [streamTitle, setStreamTitle] = useState('');
    const [watchingNow, setWatchingNow] = useState(0);
    const [createdOn, setCreatedOn] = useState('');

    useEffect(() => {
        console.log('use effect start');
        Start(streamKey);
        setActive(true);

        EventsOn('QueryResponse', (response) => {
            console.log('query response received');
            setRefresh(!refresh);
            setActive(true);
            setChannelName(response.channel_name);
            setFollowers(response.followers);
            setChannelFollowers(response.followers.num_followers);
            setTotalFollowers(response.followers.num_followers_total);
            setRecentFollowers(response.followers.recent_followers);
            setSubscribers(response.subscribers);
            setSubscriberCount(response.subscribers.num_subscribers);
            setRecentSubscribers(response.subscribers.recent_subscribers);
            if (response.livestreams.length > 0) {
                setStreamLive(true);
                setStreamCategories(response.livestreams[0].categories);
                setStreamLikes(response.livestreams[0].likes);
                setStreamDislikes(response.livestreams[0].dislikes);
                setStreamTitle(response.livestreams[0].title);
                setCreatedOn(response.livestreams[0].created_on);
                setWatchingNow(response.livestreams[0].watching_now);
            } else {
                setStreamLive(false);
            }
        });
    }, []);

    const startQuery = () => {
        console.log('start');
        Start(streamKey);
        setActive(true);
    };

    const stopQuery = () => {
        console.log('stop');
        Stop();
        // EventsEmit('StopQuery');
        setActive(false);
    };

    const activityDate = (activity) => {
        if (activity.followed_on) {
            return activity.followed_on;
        }
        if (activity.subscribed_on) {
            return activity.subscribed_on;
        }
    };

    const activityEvents = () => {
        let sorted = [...recentFollowers, ...recentSubscribers].sort((a, b) =>
            activityDate(a) < activityDate(b) ? 1 : -1
        );
        return sorted;
    };

    return (
        <div id='Dashboard'>
            <div className='header'>
                <div className='header-left'></div>
                <div className='highlights'>
                    {/* <Highlight description={'Session'} type={'stopwatch'} value={createdOn} /> */}
                    <Highlight description={'Viewers'} type={'count'} value={watchingNow} />
                    <Highlight description={'Followers'} type={'count'} value={channelFollowers} />
                    <Highlight description={'Subscribers'} type={'count'} value={subscriberCount} />
                </div>
                <div className='header-right'></div>
            </div>
            <div className='main'>
                <div className='main-left'>
                    <StreamActivity title={'Stream Activity'} events={activityEvents()} />
                </div>
                <div className='main-right'>
                    <StreamChat title={'Stream Chat'} />
                </div>
                <div></div>
            </div>
            <StreamInfo
                active={active}
                channel={channelName}
                title={streamTitle}
                categories={streamCategories}
                likes={streamLikes}
                live={streamLive}
                dislikes={streamDislikes}
                play={startQuery}
                pause={stopQuery}
            />
        </div>
    );
}

export default Dashboard;
