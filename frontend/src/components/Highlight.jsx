import './Highlight.css';

function Highlight(props) {
    const countString = () => {
        switch (true) {
            case props.value <= 0:
                return '-';
            case props.value < 1000:
                return props.value;
            case props.value < 1000000:
                return (props.value / 1000).toFixed(3).slice(0, -2) + 'K';
            case props.value < 1000000000:
                return (props.value / 1000000).toFixed(6).slice(0, -5) + 'M';
            default:
                return 'Inf';
        }
    };

    const stopwatchString = () => {
        if (isNaN(Date.parse(props.value))) {
            return '--:--';
        }
        let now = new Date();
        let date = new Date(props.value);
        let diff = now - date;

        let msMinute = 1000 * 60;
        let msHour = msMinute * 60;
        let msDay = msHour * 24;

        let days = Math.floor(diff / msDay);
        let hours = Math.floor((diff - days * msDay) / msHour);
        let minutes = Math.floor((diff - days * msDay - hours * msHour) / msMinute);

        if (diff >= 100 * msDay) {
            return days + 'd';
        }
        if (diff >= msDay) {
            return days + 'd ' + hours + 'h';
        }

        if (hours < 10) {
            hours = '0' + hours;
        }

        if (minutes < 10) {
            minutes = '0' + minutes;
        }

        return hours + ':' + minutes;
    };

    const valueString = () => {
        switch (props.type) {
            case 'count':
                return countString();
            case 'stopwatch':
                return stopwatchString();
            default:
                return props.value;
        }
    };

    return (
        <div className='highlight'>
            <span className='highlight-value'>{valueString()}</span>
            <span className='highlight-description'>{props.description}</span>
        </div>
    );
}

export default Highlight;
