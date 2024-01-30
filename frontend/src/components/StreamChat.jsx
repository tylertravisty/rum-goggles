import { useState } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import ChatMessage from './ChatMessage';
import './StreamChat.css';

function StreamChat(props) {
    const [messages, setMessages] = useState([
        {
            color: '#ec131f',
            image: 'https://ak2.rmbl.ws/z0/V/m/v/E/VmvEe.asF.4-18osof-s35kf7.jpeg',
            username: 'tylertravisty',
            text: 'Hello, world this is si s a a sdf asd f',
        },
        {
            username: 'tylertravisty',
            text: 'Another chat message',
        },
    ]);

    EventsOn('ChatMessage', (msg) => {
        setMessages(...messages, msg);
    });

    return (
        <div className='stream-chat'>
            <div className='stream-chat-header'>
                <span className='stream-chat-title'>{props.title}</span>
            </div>
            <div className='stream-chat-list'>
                {messages.map((message, index) => (
                    <ChatMessage message={message} />
                ))}
            </div>
        </div>
    );
}

export default StreamChat;
