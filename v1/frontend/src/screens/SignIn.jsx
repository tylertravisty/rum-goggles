import { useEffect, useState } from 'react';
import { SmallModal } from '../components/Modal';
import { Login, SignedIn } from '../../wailsjs/go/main/App';
import { Eye, EyeSlash, Logo } from '../assets';
import { Navigate, useNavigate } from 'react-router-dom';
import './SignIn.css';
import { NavDashboard } from '../Navigation';

function SignIn() {
    const [error, setError] = useState('');
    const navigate = useNavigate();
    const [password, setPassword] = useState('');
    const updatePassword = (event) => setPassword(event.target.value);
    const [showPassword, setShowPassword] = useState(false);
    const updateShowPassword = () => setShowPassword(!showPassword);
    const [signingIn, setSigningIn] = useState(false);
    const [username, setUsername] = useState('');
    const updateUsername = (event) => setUsername(event.target.value);

    // useEffect(() => {
    //     SignedIn()
    //         .then((signedIn) => {
    //             if (signedIn) {
    //                 navigate(NavDashboard);
    //             }
    //         })
    //         .catch((error) => {
    //             setError(error);
    //         });
    // }, []);

    useEffect(() => {
        if (signingIn) {
            Login(username, password)
                .then(() => {
                    setUsername('');
                    setPassword('');
                    navigate(NavDashboard);
                })
                .catch((error) => {
                    setError(error);
                })
                .finally(() => {
                    setSigningIn(false);
                });
        }
    }, [signingIn]);

    const signIn = () => {
        setSigningIn(true);
    };

    return (
        <>
            <SmallModal
                onClose={() => setError('')}
                show={error !== ''}
                style={{ minWidth: '300px', maxWidth: '200px', maxHeight: '200px' }}
                title={'Error'}
                message={error}
                submitButton={'OK'}
                onSubmit={() => setError('')}
            />
            <div className='signin-body'>
                <div className='signin-header'>
                    <img className='signin-logo' src={Logo} />
                </div>
                <div></div>
                <div className='signin-center'>
                    <div className='signin-window'>
                        <div className='signin-window-header'>
                            <span className='signin-window-title'>Sign in to Rumble</span>
                        </div>
                        <div className='signin-window-form'>
                            <div className='signin-window-field'>
                                <span className='signin-window-field-label'>Username</span>
                                <input
                                    className='signin-window-field-input'
                                    onChange={updateUsername}
                                    placeholder='Username'
                                    value={username}
                                />
                            </div>
                            <div className='signin-window-field'>
                                <span className='signin-window-field-label'>Password</span>
                                <div className='signin-window-password'>
                                    <input
                                        className='signin-window-password-input'
                                        onChange={updatePassword}
                                        placeholder='Password'
                                        type={showPassword ? 'text' : 'password'}
                                        value={password}
                                    />
                                    <button
                                        className='signin-window-password-show-button'
                                        onClick={updateShowPassword}
                                    >
                                        <img
                                            className='signin-window-password-show-icon'
                                            src={showPassword ? EyeSlash : Eye}
                                        />
                                    </button>
                                </div>
                            </div>
                            <button className='signin-window-form-button' onClick={signIn}>
                                {signingIn ? 'Signing in...' : 'Sign In'}
                            </button>
                        </div>
                        <div></div>
                    </div>
                </div>
                <div className='signin-footer'>
                    <span className='signin-footer-description'>Rum Goggles by Tyler Travis</span>
                    <span className='signin-footer-description'>Follow @tylertravisty</span>
                </div>
            </div>
        </>
    );
}

export default SignIn;
