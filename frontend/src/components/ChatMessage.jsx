import './ChatMessage.css';

function ChatMessage(props) {
    const upperInitial = () => {
        return props.message.username[0].toUpperCase();
    };

    return (
        <div className='chat-message'>
            {props.message.image === '' || props.message.image === undefined ? (
                <span className='chat-message-user-initial'>{upperInitial()}</span>
            ) : (
                <img className='chat-message-user-image' src={props.message.image} />
            )}
            <div>
                <span
                    className='chat-message-username'
                    style={props.message.color && { color: props.message.color }}
                >
                    {props.message.username}
                </span>
                <span className='chat-message-text'>{props.message.text}</span>
            </div>
        </div>
    );
}

export default ChatMessage;
