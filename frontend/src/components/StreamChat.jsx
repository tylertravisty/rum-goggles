import './StreamChat.css';

function StreamChat(props) {
    return (
        <div className='stream-chat'>
            <div className='stream-chat-header'>
                <span className='stream-chat-title'>{props.title}</span>
            </div>
        </div>
    );
}

export default StreamChat;
