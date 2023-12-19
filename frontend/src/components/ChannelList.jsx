import { PlusCircle } from '../assets/icons';
import './ChannelList.css';

function ChannelList(props) {
    const sortChannelsAlpha = () => {
        let sorted = [...props.channels].sort((a, b) =>
            a.name.toLowerCase() > b.name.toLowerCase() ? 1 : -1
        );
        return sorted;
    };

    return (
        <div className='channel-list'>
            <span className='channel-list-title'>Channels</span>
            <div className='channels'>
                {sortChannelsAlpha().map((channel, index) => (
                    <div className='channel' style={index === 0 ? { borderTop: 'none' } : {}}>
                        <button
                            className='channel-button'
                            onClick={() => props.openStreamDashboard(channel.api_url)}
                        >
                            {channel.name}
                        </button>
                    </div>
                ))}
            </div>
            {/* <button className='channel-add'>
                <img className='channel-add-icon' src={PlusCircle} />
            </button> */}
        </div>
    );
}

export default ChannelList;
