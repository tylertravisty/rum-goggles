import { useEffect, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { NavDashboard } from './Navigation';
import { AddChannel, Config } from '../../wailsjs/go/main/App';
import { Eye, EyeSlash } from '../assets/icons';
import './SignIn.css';
import ChannelList from '../components/ChannelList';
import { SmallModal } from '../components/Modal';

function SignIn() {
    const [error, setError] = useState('');
    const navigate = useNavigate();
    const [config, setConfig] = useState({ channels: {} });
    const [addChannelError, setAddChannelError] = useState('');
    const [streamKey, setStreamKey] = useState('');
    const updateStreamKey = (event) => setStreamKey(event.target.value);
    const [showStreamKey, setShowStreamKey] = useState(false);
    const updateShowStreamKey = () => setShowStreamKey(!showStreamKey);

    useEffect(() => {
        Config()
            .then((response) => {
                setConfig(response);
            })
            .catch((error) => {
                // TODO: display error to user
                setError('Error loading config: ' + error);
                console.log('error getting config', error);
            });
    }, []);

    const saveStreamKey = () => {
        AddChannel(streamKey)
            .then((response) => {
                console.log(response);
                setConfig(response);
                setStreamKey('');
            })
            .catch((error) => {
                console.log('error adding channel', error);
                setAddChannelError(error);
            });
    };

    const openStreamDashboard = (cid) => {
        navigate(NavDashboard, { state: { cid: cid } });
    };

    return (
        <>
            {error !== '' && (
                <SmallModal
                    onClose={() => setError('')}
                    show={error !== ''}
                    style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                    title={'Error'}
                    message={error}
                    submitButton={'OK'}
                    onSubmit={() => setError('')}
                />
            )}
            <div id='SignIn'>
                <div className='signin-header'>
                    <span className='signin-title-text'>Rum Goggles</span>
                    <span className='signin-title-subtext'>Rumble Stream Dashboard</span>
                </div>
                <div className='signin-center'>
                    <ChannelList
                        channels={config.channels}
                        openStreamDashboard={openStreamDashboard}
                    />
                </div>
                <div className='signin-input-box'>
                    <label className='signin-label'>Add Channel</label>
                    <span className='add-channel-description'>
                        Copy your API key from your Rumble account
                    </span>
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
                    <span className='add-channel-error'>
                        {addChannelError ? addChannelError : '\u00A0'}
                    </span>
                </div>
                <div className='signin-footer'></div>
            </div>
        </>
    );
}

export default SignIn;
