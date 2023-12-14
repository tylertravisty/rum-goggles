import { useEffect, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { NavDashboard } from './Navigation';
import { Eye, EyeSlash } from '../assets/icons';
import './SignIn.css';

function SignIn() {
    const navigate = useNavigate();
    const [streamKey, setStreamKey] = useState('');
    const updateStreamKey = (event) => setStreamKey(event.target.value);
    const [showStreamKey, setShowStreamKey] = useState(false);
    const updateShowStreamKey = () => setShowStreamKey(!showStreamKey);

    const saveStreamKey = () => {
        navigate(NavDashboard, { state: { streamKey: streamKey } });
    };

    return (
        <div id='SignIn'>
            <div className='signin-title'>
                <span className='signin-title-text'>Rum Goggles</span>
                <span className='signin-title-subtext'>Rumble Stream Dashboard</span>
            </div>
            <div className='signin-input-box'>
                <label className='signin-label'>Stream Key:</label>
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
            <div className='signin-title'></div>
        </div>
    );
}

export default SignIn;
