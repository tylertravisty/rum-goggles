import { useEffect, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { NavDashboard } from './Navigation';
import { AddChannel, Config } from '../../wailsjs/go/main/App';
import { Eye, EyeSlash } from '../assets/icons';
import './SignIn.css';
import ChannelList from '../components/ChannelList';

function SignIn() {
    const navigate = useNavigate();
    const [config, setConfig] = useState({ channels: {} });
    const [streamKey, setStreamKey] = useState('');
    const updateStreamKey = (event) => setStreamKey(event.target.value);
    const [showStreamKey, setShowStreamKey] = useState(false);
    const updateShowStreamKey = () => setShowStreamKey(!showStreamKey);

    useEffect(() => {
        Config()
            .then((response) => {
                setConfig(response);
            })
            .catch((err) => {
                // TODO: display error to user
                console.log('error getting config', err);
            });
    }, []);

    const saveStreamKey = () => {
        AddChannel(streamKey)
            .then((response) => {
                console.log(response);
                setConfig(response);
                setStreamKey('');
            })
            .catch((err) => {
                console.log('error adding channel', err);
            });
    };

    const openStreamDashboard = (cid) => {
        navigate(NavDashboard, { state: { cid: cid } });
    };

    return (
        <div id='SignIn'>
            <div className='signin-header'>
                <span className='signin-title-text'>Rum Goggles</span>
                <span className='signin-title-subtext'>Rumble Stream Dashboard</span>
            </div>
            <div className='signin-center'>
                <ChannelList channels={config.channels} openStreamDashboard={openStreamDashboard} />
            </div>
            <div className='signin-input-box'>
                <label className='signin-label'>Add Channel</label>
                <div className='signin-input-button'>
                    <input
                        id='StreamKey'
                        className='signin-input'
                        onChange={updateStreamKey}
                        placeholder='Stream Key'
                        type={showStreamKey ? 'text' : 'password'}
                        value={streamKey}
                    />
                    <button className='signin-show' onClick={updateShowStreamKey}>
                        <img
                            className='signin-show-icon'
                            src={showStreamKey ? EyeSlash : Eye}
                        ></img>
                    </button>
                    <button className='signin-button' onClick={saveStreamKey}>
                        Save
                    </button>
                </div>
            </div>
            <div className='signin-footer'></div>
        </div>
    );
}

export default SignIn;
