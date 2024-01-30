import { useEffect, useState } from 'react';
import { Navigate, useLocation, useNavigate } from 'react-router-dom';
import {
    AddChatMessage,
    ChatBotMessages,
    DeleteChatMessage,
    NewChatBot,
    ResetChatBot,
    StartApi,
    StopApi,
    StopChatBotMessage,
    UpdateChatMessage,
} from '../../wailsjs/go/main/App';

import './Dashboard.css';
import { EventsEmit, EventsOn } from '../../wailsjs/runtime/runtime';
import { Heart, Star } from '../assets/icons';
import { ChatBotModal } from '../components/ChatBot';
import Highlight from '../components/Highlight';
import { SmallModal } from '../components/Modal';
import StreamEvent from '../components/StreamEvent';
import StreamActivity from '../components/StreamActivity';
import StreamChat from '../components/StreamChat';
import StreamChatBot from '../components/StreamChatBot';
import StreamInfo from '../components/StreamInfo';
import { NavSignIn } from './Navigation';
import { StreamChatMessageModal } from '../components/StreamChatMessage';

function Dashboard() {
    const location = useLocation();
    const navigate = useNavigate();
    const [error, setError] = useState('');
    const [refresh, setRefresh] = useState(false);
    const [active, setActive] = useState(false);
    const [openChatBot, setOpenChatBot] = useState(false);
    const [chatBotMessages, setChatBotMessages] = useState({});
    const [chatAsChannel, setChatAsChannel] = useState(false);
    const [chatCommand, setChatCommand] = useState('');
    const [chatOnCommand, setChatOnCommand] = useState(false);
    const [chatID, setChatID] = useState('');
    const [chatInterval, setChatInterval] = useState('');
    const [chatText, setChatText] = useState('');
    const [chatTextFile, setChatTextFile] = useState('');
    const [openChat, setOpenChat] = useState(false);
    const [cid, setCID] = useState(location.state.cid);
    const [username, setUsername] = useState('');
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
        // TODO: catch error
        StartApi(cid);
        setActive(true);

        ChatBotMessages(cid).then((messages) => {
            console.log(messages);
            setChatBotMessages(messages);
        });

        EventsOn('QueryResponse', (response) => {
            console.log('query response received');
            setRefresh(!refresh);
            setActive(true);
            setUsername(response.username);
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

        EventsOn('QueryResponseError', (error) => {
            setError(error);
            console.log('Query response error:', error);
            setActive(false);
        });
    }, []);

    const home = () => {
        StopApi()
            .then(() => setActive(false))
            .then(() => {
                ResetChatBot();
            })
            .then(() => {
                navigate(NavSignIn);
            })
            .catch((error) => {
                setError(error);
                console.log('Stop error:', error);
            });
    };

    const startQuery = () => {
        console.log('start');
        StartApi(cid)
            .then(() => {
                setActive(true);
            })
            .catch((error) => {
                setError(error);
                console.log('Start error:', error);
            });
    };

    const stopQuery = () => {
        console.log('stop');
        StopApi().then(() => {
            setActive(false);
        });
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

    const newChat = () => {
        setChatAsChannel(false);
        setChatCommand('');
        setChatID('');
        setChatInterval('');
        setChatText('');
        setChatTextFile('');
        setChatOnCommand(false);
        setOpenChat(true);
    };

    const editChat = (id, asChannel, command, interval, onCommand, text, textFile) => {
        setChatAsChannel(asChannel);
        setChatCommand(command);
        setChatID(id);
        setChatInterval(interval);
        setChatOnCommand(onCommand);
        setChatText(text);
        setChatTextFile(textFile);
        setOpenChat(true);
    };

    const deleteChat = (id) => {
        setOpenChat(false);
        if (id === '') {
            return;
        }

        StopChatBotMessage(id, cid)
            .then(() => {
                DeleteChatMessage(id, cid)
                    .then((messages) => {
                        setChatBotMessages(messages);
                    })
                    .catch((error) => {
                        setError(error);
                        console.log('Error deleting message:', error);
                    });
            })
            .catch((error) => {
                setError(error);
                console.log('Error stopping message:', error);
            });
    };

    const saveChat = (id, asChannel, command, interval, onCommand, text, textFile) => {
        console.log('save chat textfile:', textFile);
        setOpenChat(false);
        if (id === '') {
            AddChatMessage(cid, asChannel, command, interval, onCommand, text, textFile)
                .then((messages) => {
                    setChatBotMessages(messages);
                })
                .catch((error) => {
                    setError(error);
                    console.log('Error saving chat:', error);
                });

            return;
        }

        UpdateChatMessage(id, cid, asChannel, command, interval, onCommand, text, textFile)
            .then((messages) => {
                console.log(messages);
                setChatBotMessages(messages);
            })
            .catch((error) => {
                setError(error);
                console.log('Error saving chat:', error);
            });
    };

    const saveChatBot = (username, password, url) => {
        NewChatBot(cid, username, password, url)
            .then(() => {
                setOpenChatBot(false);
            })
            .catch((error) => console.log('Error creating new chat bot:', error));
    };

    return (
        <>
            {openChat && (
                <StreamChatMessageModal
                    chatID={chatID}
                    asChannel={chatAsChannel}
                    chatCommand={chatCommand}
                    onCommand={chatOnCommand}
                    interval={chatInterval}
                    onClose={() => setOpenChat(false)}
                    onDelete={deleteChat}
                    onSubmit={saveChat}
                    show={openChat}
                    text={chatText}
                    textFile={chatTextFile}
                />
            )}
            {openChatBot && (
                <ChatBotModal
                    cid={cid}
                    onClose={() => setOpenChatBot(false)}
                    onSubmit={saveChatBot}
                    show={openChatBot}
                />
            )}
            <div id='Dashboard'>
                <div className='header'>
                    <div className='header-left'></div>
                    <div className='highlights'>
                        {/* <Highlight description={'Session'} type={'stopwatch'} value={createdOn} /> */}
                        <Highlight description={'Viewers'} type={'count'} value={watchingNow} />
                        <Highlight
                            description={'Followers'}
                            type={'count'}
                            value={channelFollowers}
                        />
                        <Highlight
                            description={'Subscribers'}
                            type={'count'}
                            value={subscriberCount}
                        />
                    </div>
                    <div className='header-right'></div>
                </div>
                <div className='main'>
                    <div className='main-left'>
                        <StreamActivity title={'Stream Activity'} events={activityEvents()} />
                    </div>
                    {/* <div className='main-middle'>
                        <StreamChat title={'Stream Chat'} />
                    </div> */}
                    <div className='main-right'>
                        <StreamChatBot
                            chats={chatBotMessages}
                            onAdd={newChat}
                            onEdit={editChat}
                            onSettings={() => setOpenChatBot(true)}
                            title={'Chat Bot'}
                        />
                    </div>
                </div>
                <StreamInfo
                    active={active}
                    channel={channelName !== '' ? channelName : username}
                    title={streamTitle}
                    categories={streamCategories}
                    likes={streamLikes}
                    live={streamLive}
                    dislikes={streamDislikes}
                    home={home}
                    play={startQuery}
                    pause={stopQuery}
                    // settings={openModal}
                />
            </div>
            {error !== '' && (
                <SmallModal
                    onClose={() => setError('')}
                    show={error !== ''}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Error'}
                    message={error}
                    submitButton={'OK'}
                    onSubmit={() => setError('')}
                />
            )}
        </>
    );
}

export default Dashboard;
