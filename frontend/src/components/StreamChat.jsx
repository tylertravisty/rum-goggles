import { useEffect, useState } from 'react';
import { StartChatBotMessage, StopChatBotMessage } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GearFill, Pause, Play, PlusCircle } from '../assets/icons';
import './StreamChat.css';
import { SmallModal } from './Modal';

function StreamChat(props) {
    const sortChatsAlpha = () => {
        let keys = Object.keys(props.chats);

        let sorted = [...keys].sort((a, b) =>
            props.chats[a].text.toLowerCase() > props.chats[b].text.toLowerCase() ? 1 : -1
        );
        return sorted;
    };

    return (
        <div className='stream-chat'>
            <div className='stream-chat-header'>
                <span className='stream-chat-title'>{props.title}</span>
                <div className='stream-chat-controls'>
                    <button
                        className='stream-chat-button stream-chat-button-title'
                        onClick={props.onAdd}
                    >
                        <img className='stream-chat-icon' src={PlusCircle} />
                    </button>
                    <button
                        className='stream-chat-button stream-chat-button-title'
                        onClick={props.onSettings}
                    >
                        <img className='stream-chat-icon' src={GearFill} />
                    </button>
                </div>
            </div>
            <div className='stream-chat-list'>
                {sortChatsAlpha().map((chat, index) => (
                    <StreamChatItem chat={props.chats[chat]} onItemClick={props.onEdit} />
                ))}
            </div>
        </div>
    );
}

export default StreamChat;

function StreamChatItem(props) {
    const [active, setActive] = useState(props.chat.active);
    const [error, setError] = useState('');

    const changeActive = (bool) => {
        console.log('ChangeActive:', bool);
        props.chat.active = bool;
        setActive(bool);
    };

    useEffect(() => {
        EventsOn('ChatBotMessageActive-' + props.chat.id, (mid) => {
            console.log('ChatBotMessageActive', props.chat.id, mid);
            if (mid === props.chat.id) {
                changeActive(true);
            }
        });

        EventsOn('ChatBotMessageError-' + props.chat.id, (mid) => {
            console.log('ChatBotMessageError', props.chat.id, mid);
            if (mid === props.chat.id) {
                changeActive(false);
            }
        });
    }, []);

    const prependZero = (value) => {
        if (value < 10) {
            return '0' + value;
        }

        return '' + value;
    };

    const printInterval = (interval) => {
        let hours = Math.floor(interval / 3600);
        let minutes = Math.floor(interval / 60 - hours * 60);
        let seconds = Math.floor(interval - hours * 3600 - minutes * 60);

        // hours = prependZero(hours);
        // minutes = prependZero(minutes);
        // seconds = prependZero(seconds);
        // return hours + ':' + minutes + ':' + seconds;

        return hours + 'h ' + minutes + 'm ' + seconds + 's';
    };

    const intervalToTimer = (interval) => {
        let hours = Math.floor(interval / 3600);
        let minutes = Math.floor(interval / 60 - hours * 60);
        let seconds = Math.floor(interval - hours * 3600 - minutes * 60);

        if (minutes !== 0) {
            seconds = prependZero(seconds);
        }
        if (hours !== 0) {
            minutes = prependZero(minutes);
        }
        if (hours === 0) {
            hours = '';
            if (minutes === 0) {
                minutes = '';
                if (seconds === 0) {
                    seconds = '';
                }
            }
        }

        return hours + minutes + seconds;
    };

    const openChat = () => {
        props.onItemClick(
            props.chat.id,
            props.chat.as_channel,
            intervalToTimer(props.chat.interval),
            props.chat.text
        );
    };

    const startMessage = () => {
        StartChatBotMessage(props.chat.id)
            .then(() => {
                changeActive(true);
            })
            .catch((error) => {
                setError(error);
            });
    };

    const stopMessage = () => {
        StopChatBotMessage(props.chat.id).then(() => {
            changeActive(false);
        });
    };

    return (
        <>
            <SmallModal
                onClose={() => setError('')}
                show={error !== ''}
                style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                title={'Error'}
                message={error}
                submitButton={'OK'}
                onSubmit={() => setError('')}
            />
            <div className='stream-chat-item' onClick={() => openChat()}>
                <span className='stream-chat-item-message'>{props.chat.text}</span>
                <span className='stream-chat-item-interval'>
                    {printInterval(props.chat.interval)}
                </span>
                <span className='stream-chat-item-sender'>
                    {props.chat.as_channel ? 'Channel' : 'User'}
                </span>
                <button
                    className='stream-chat-button stream-chat-button-chat'
                    onClick={(e) => {
                        e.stopPropagation();
                        console.log('message ID:', props.chat.id);
                        if (active) {
                            console.log('Stop message');
                            stopMessage();
                        } else {
                            console.log('Start message');
                            startMessage();
                        }
                    }}
                >
                    <img className='stream-chat-icon' src={active ? Pause : Play} />
                </button>
            </div>
        </>
    );
}
