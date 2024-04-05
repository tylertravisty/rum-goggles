import { useEffect, useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { NavDashboard, NavSignIn } from '../Navigation';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { Start } from '../../wailsjs/go/main/App';
import { Logo } from '../assets';
import './Startup.css';

function Startup() {
    const [error, setError] = useState('');
    const [message, setMessage] = useState('');
    const navigate = useNavigate();
    const [starting, setStarting] = useState(true);

    useEffect(() => {
        EventsOn('StartupMessage', (event) => {
            setMessage(event);
        });
        setStarting(false);
    }, []);

    useEffect(() => {
        if (!starting) {
            Start()
                .then((signin) => {
                    if (signin) {
                        navigate(NavSignIn);
                    } else {
                        navigate(NavDashboard);
                    }
                })
                .catch((err) => {
                    setError(err);
                });
        }
    }, [starting]);

    return (
        <div className='startup-body'>
            <div className='startup-header'>
                <img className='startup-logo' src={Logo} />
            </div>
            <div className='startup-center'>
                <div className='startup-window'>
                    <div className='startup-window-header'>
                        <span className='startup-window-title'>Starting Rum Goggles</span>
                    </div>
                    <div className='startup-window-message'>
                        {error !== '' && <span className='startup-error'>{error}</span>}
                        {message !== '' && error === '' && (
                            <span className='startup-message'>{message}</span>
                        )}
                    </div>
                </div>
            </div>
            <div className='startup-footer'>
                <span className='startup-footer-description'>Rum Goggles by Tyler Travis</span>
                <span className='startup-footer-description'>Follow @tylertravisty</span>
            </div>
        </div>
    );
}

export default Startup;
