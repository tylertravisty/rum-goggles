import { Heart, Star } from '../assets/icons';

import './StreamEvent.css';

function StreamEvent(props) {
    const dateDate = (date) => {
        const options = { month: 'short' };
        let month = new Intl.DateTimeFormat('en-US', options).format(date);
        let day = date.getDate();
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
        if (minutes < 10) {
            minutes = '0' + minutes;
        }

        let mer = 'pm';
        if (hours24 < 12) {
            mer = 'am';
        }

        return hours + ':' + minutes + ' ' + mer;
    };

    const dateString = (d) => {
        if (isNaN(Date.parse(d))) {
            return 'Who knows?';
        }

        let now = new Date();
        let date = new Date(d);
        // Fix Rumble's timezone problem
        date.setHours(date.getHours() - 4);
        let diff = now - date;
        switch (true) {
            case diff < 0:
                return 'In the future!?';
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
    };

    return (
        <div className='stream-event'>
            <div className='stream-event-left'>
                {props.event.followed_on && <img className='stream-event-icon' src={Heart}></img>}
                {props.event.subscribed_on && <img className='stream-event-icon' src={Star}></img>}
                <div className='stream-event-left-text'>
                    <span className='stream-event-username'>{props.event.username}</span>
                    <span className='stream-event-description'>
                        {props.event.followed_on && 'Followed you'}
                        {props.event.subscribed_on && 'Subscribed'}
                    </span>
                </div>
            </div>
            <span className='stream-event-date'>
                {props.event.followed_on && dateString(props.event.followed_on)}
                {props.event.subscribed_on && dateString(props.event.subscribed_on)}
            </span>
        </div>
    );
}

export default StreamEvent;
