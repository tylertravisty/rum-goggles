import { PlusCircle } from '../assets/icons';
import './StreamChat.css';

function StreamChat(props) {
    return (
        <div className='stream-chat'>
            <div className='stream-chat-header'>
                <span className='stream-chat-title'>{props.title}</span>
                <button
                    onClick={() => console.log('Add chat bot')}
                    className='stream-chat-add-button'
                >
                    <img className='stream-chat-add-icon' src={PlusCircle} />
                </button>
            </div>
        </div>
    );
}

export default StreamChat;
