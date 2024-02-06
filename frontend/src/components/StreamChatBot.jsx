import { useEffect, useState } from 'react';
import { FilepathBase, StartChatBotMessage, StopChatBotMessage } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GearFill, Pause, Play, PlusCircle } from '../assets/icons';
import './StreamChatBot.css';
import { SmallModal } from './Modal';

function StreamChatBot(props) {
    const sortChatsAlpha = () => {
        let keys = Object.keys(props.chats);

        let sorted = [...keys].sort((a, b) =>
            props.chats[a].text.toLowerCase() > props.chats[b].text.toLowerCase() ? 1 : -1
        );
        return sorted;
    };

    return (
        <div className='stream-chatbot'>
            <div className='stream-chatbot-header'>
                <span className='stream-chatbot-title'>{props.title}</span>
                <div className='stream-chatbot-controls'>
                    <button
                        className='stream-chatbot-button stream-chatbot-button-title'
                        onClick={props.onAdd}
                    >
                        <img className='stream-chatbot-icon' src={PlusCircle} />
                    </button>
                    <button
                        className='stream-chatbot-button stream-chatbot-button-title'
                        onClick={props.onSettings}
                    >
                        <img className='stream-chatbot-icon' src={GearFill} />
                    </button>
                </div>
            </div>
            <div className='stream-chatbot-list'>
                {sortChatsAlpha().map((chat, index) => (
                    <StreamChatItem
                        activateMessage={props.activateMessage}
                        chat={props.chats[chat]}
                        isMessageActive={props.isMessageActive}
                        onItemClick={props.onEdit}
                    />
                ))}
            </div>
        </div>
    );
}

export default StreamChatBot;

function StreamChatItem(props) {
    const [active, setActive] = useState(props.isMessageActive(props.chat.id));
    const [error, setError] = useState('');
    const [filename, setFilename] = useState(props.chat.text_file);

    useEffect(() => {
        if (props.chat.text_file !== '') {
            FilepathBase(props.chat.text_file).then((name) => {
                setFilename(name);
            });
        }
        setActive(props.isMessageActive(props.chat.id));
    }, [props]);

    const changeActive = (bool) => {
        // console.log('ChangeActive:', bool);
        // props.chat.active = bool;
        props.activateMessage(props.chat.id, bool);
        setActive(bool);
    };

    useEffect(() => {
        EventsOn('ChatBotCommandActive-' + props.chat.id, (mid) => {
            console.log('ChatBotCommandActive', props.chat.id, mid);
            if (mid === props.chat.id) {
                changeActive(true);
            }
        });

        EventsOn('ChatBotCommandError-' + props.chat.id, (mid) => {
            console.log('ChatBotCommandError', props.chat.id, mid);
            if (mid === props.chat.id) {
                changeActive(false);
            }
        });

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
        props.onItemClick({
            id: props.chat.id,
            as_channel: props.chat.as_channel,
            command: props.chat.command,
            interval: intervalToTimer(props.chat.interval),
            on_command: props.chat.on_command,
            on_command_follower: props.chat.on_command_follower,
            on_command_rant_amount: props.chat.on_command_rant_amount,
            on_command_subscriber: props.chat.on_command_subscriber,
            text: props.chat.text,
            text_file: props.chat.text_file,
        });
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
            <div className='stream-chatbot-item' onClick={() => openChat()}>
                <span className='stream-chatbot-item-message'>
                    {props.chat.text_file !== '' ? filename : props.chat.text}
                </span>
                <span className='stream-chatbot-item-interval'>
                    {props.chat.on_command
                        ? props.chat.command
                        : printInterval(props.chat.interval)}
                </span>
                <span className='stream-chatbot-item-sender'>
                    {props.chat.as_channel ? 'Channel' : 'User'}
                </span>
                <button
                    className='stream-chatbot-button stream-chatbot-button-chat'
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
                    <img className='stream-chatbot-icon' src={active ? Pause : Play} />
                </button>
            </div>
        </>
    );
}
