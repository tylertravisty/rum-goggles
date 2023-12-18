import StreamEvent from './StreamEvent';

import './StreamActivity.css';

function StreamActivity(props) {
    return (
        <div className='stream-activity'>
            <div className='stream-activity-header'>
                <span className='stream-activity-title'>{props.title}</span>
            </div>
            <div className='stream-activity-list'>
                {props.events.map((event, index) => (
                    <StreamEvent event={event} />
                ))}
            </div>
        </div>
    );
}

export default StreamActivity;
