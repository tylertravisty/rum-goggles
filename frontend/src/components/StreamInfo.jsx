import { Gear, House, Pause, Play, ThumbsDown, ThumbsUp } from '../assets/icons';
import './StreamInfo.css';

function StreamInfo(props) {
    const likesString = (likes) => {
        switch (true) {
            case likes <= 0:
                return '0';
            case likes < 1000:
                return likes;
            case likes < 1000000:
                return (likes / 1000).toFixed(3).slice(0, -2) + 'K';
            case likes < 1000000000:
                return (likes / 1000000).toFixed(6).slice(0, -5) + 'M';
            default:
                return 'Inf';
        }
    };

    return (
        <div className='stream-info'>
            <div className='stream-info-live'>
                <div className='stream-info-title'>
                    <span>{props.live ? props.title : '-'}</span>
                </div>
                <div className='stream-info-subtitle'>
                    <div className='stream-info-categories'>
                        <span className='stream-info-category'>
                            {props.live ? props.categories.primary.title : 'none'}
                        </span>
                        <span className='stream-info-category'>
                            {props.live ? props.categories.secondary.title : 'none'}
                        </span>
                    </div>
                    <div className='stream-info-likes'>
                        <div className='stream-info-likes-left'>
                            <img className='stream-info-likes-icon' src={ThumbsUp} />
                            <span className='stream-info-likes-count'>
                                {props.live ? likesString(props.likes) : '-'}
                            </span>
                        </div>
                        <div className='stream-info-likes-right'>
                            <img className='stream-info-likes-icon' src={ThumbsDown} />
                            <span className='stream-info-likes-count'>
                                {props.live ? likesString(props.dislikes) : '-'}
                            </span>
                        </div>
                    </div>
                </div>
            </div>
            <div className='stream-info-channel'>
                <span>Channel: {props.channel}</span>
            </div>
            <div className='stream-info-footer'>
                <div></div>
                <div className='stream-info-controls'>
                    <button className='stream-info-control-button'>
                        <img className='stream-info-control' src={House} />
                    </button>
                    <button className='stream-info-control-button'>
                        <img
                            onClick={props.active ? props.pause : props.play}
                            className='stream-info-control'
                            src={props.active ? Pause : Play}
                        />
                    </button>
                    <button className='stream-info-control-button'>
                        <img className='stream-info-control' src={Gear} />
                    </button>
                </div>
                <div></div>
            </div>
        </div>
    );
}

export default StreamInfo;
